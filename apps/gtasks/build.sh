#!/bin/bash
# Copyright 2026 Google LLC
# Licensed under the Apache License, Version 2.0

set -e

# Resolve the execution directory. If run via 'bazel run', jump to the real source directory in the workspace.
if [ -n "$BUILD_WORKSPACE_DIRECTORY" ]; then
    cd "$BUILD_WORKSPACE_DIRECTORY/apps/gtasks"
else
    # Fallback to the script's directory
    DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    cd "$DIR"
fi

echo "========================================================================="
echo " Building GTasks Android Gateway Application..."
echo "========================================================================="

# 1. Self-healing check: Initialize Gradle wrapper if missing
if [ ! -f "./gradlew" ] || [ ! -f "gradle/wrapper/gradle-wrapper.jar" ]; then
    echo "[GTasks] Gradle wrapper files not found. Bootstrapping wrapper from official Gradle repository..."
    
    mkdir -p gradle/wrapper
    
    # Download official wrapper scripts and bootstrap jar
    echo "[GTasks] Downloading gradlew script..."
    curl -sSL -o ./gradlew https://raw.githubusercontent.com/gradle/gradle/v8.4.0/gradlew
    chmod +x ./gradlew
    
    echo "[GTasks] Downloading gradle-wrapper.jar..."
    curl -sSL -o gradle/wrapper/gradle-wrapper.jar https://raw.githubusercontent.com/gradle/gradle/v8.4.0/gradle/wrapper/gradle-wrapper.jar
    
    # Write wrapper properties file
    echo "[GTasks] Writing gradle-wrapper.properties..."
    cat <<EOF > gradle/wrapper/gradle-wrapper.properties
distributionBase=GRADLE_USER_HOME
distributionPath=wrapper/dists
distributionUrl=https\://services.gradle.org/distributions/gradle-8.4-bin.zip
zipStoreBase=GRADLE_USER_HOME
zipStorePath=wrapper/dists
EOF
    echo "[GTasks] Gradle wrapper successfully bootstrapped!"
fi

# 2. Grant execution permissions to gradlew
chmod +x ./gradlew

# 3. Compile the application
echo "[GTasks] Running Gradle compilation task..."
./gradlew assembleDebug

echo "========================================================================="
echo " Build Success! APK generated at:"
echo " $DIR/app/build/outputs/apk/debug/app-debug.apk"
echo "========================================================================="
