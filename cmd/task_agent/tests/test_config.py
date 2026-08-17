# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import os
from unittest import mock
from config import load_config, deep_merge

def test_deep_merge():
    dict1 = {"server": {"port": "8080", "address": "127.0.0.1"}}
    dict2 = {"server": {"port": "9090"}}
    merged = deep_merge(dict1, dict2)
    assert merged["server"]["port"] == "9090"
    assert merged["server"]["address"] == "127.0.0.1"

def test_load_config_port_override():
    # Verify that the PORT environment variable takes precedence for Cloud Run compatibility
    with mock.patch.dict(os.environ, {"PORT": "9999"}):
        cfg = load_config()
        assert cfg["server"]["port"] == "9999"

def test_load_config_runtime_dev():
    # Verify that MODENV_RUNTIME = dev correctly merges dev configuration properties
    with mock.patch.dict(os.environ, {"MODENV_RUNTIME": "dev"}):
        cfg = load_config()
        assert "gemini-task-engine-dev" in cfg["server"]["mcp_url"]

def test_decrypt_secret():
    from config import decrypt_secret
    plain = "my_password"
    key = b"modenv-default-key"
    encoded = "xor:" + bytes(ord(c) ^ key[i % len(key)] for i, c in enumerate(plain)).hex()
    assert decrypt_secret(encoded) == "my_password"

