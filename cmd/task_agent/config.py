import os
import tomllib
from typing import Any

def deep_merge(dict1: dict[str, Any], dict2: dict[str, Any]) -> dict[str, Any]:
    """Recursively merges dict2 into dict1."""
    merged = dict1.copy()
    for key, value in dict2.items():
        if key in merged and isinstance(merged[key], dict) and isinstance(value, dict):
            merged[key] = deep_merge(merged[key], value)
        else:
            merged[key] = value
    return merged

import binascii

def decrypt_secret(encoded_text: str) -> str:
    if not encoded_text.startswith("xor:"):
        return encoded_text
    hex_str = encoded_text[4:]
    try:
        data = binascii.unhexlify(hex_str)
    except Exception:
        return encoded_text
    key = os.environ.get("MODENV_KEY", "modenv-default-key").encode("utf-8")
    output = bytearray(len(data))
    for i in range(len(data)):
        output[i] = data[i] ^ key[i % len(key)]
    return output.decode("utf-8", errors="ignore")

def decrypt_config(val: Any) -> Any:
    if isinstance(val, str):
        return decrypt_secret(val)
    elif isinstance(val, dict):
        return {k: decrypt_config(v) for k, v in val.items()}
    elif isinstance(val, list):
        return [decrypt_config(item) for item in val]
    return val

def load_config() -> dict[str, Any]:
    # Look for configs relative to this file's directory
    current_dir = os.path.dirname(os.path.abspath(__file__))
    
    config = {}
    
    # 1. Load base configuration (.env.toml)
    base_file = os.path.join(current_dir, ".env.toml")
    if os.path.exists(base_file):
        with open(base_file, "rb") as f:
            config = tomllib.load(f)
            
    # 2. Load runtime-specific override
    runtime = os.environ.get("MODENV_RUNTIME", "local")
    override_file = os.path.join(current_dir, f".env.{runtime}.toml")
    if os.path.exists(override_file):
        with open(override_file, "rb") as f:
            override_config = tomllib.load(f)
        config = deep_merge(config, override_config)
            
    # PORT override for Cloud Run
    port = os.environ.get("PORT")
    if port:
        if "server" not in config:
            config["server"] = {}
        config["server"]["port"] = port
        
    return decrypt_config(config)


# Load settings on import
settings = load_config()
