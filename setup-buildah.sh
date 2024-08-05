#!/bin/bash

# Create necessary directories
mkdir -p containers run

# Step 2: Build the Podman Image
podman build -t buildah-container -f Containerfile-debug-buildah

# Step 3: Create the Shell Script to Run Buildah Commands
cat << 'EOF' > buildah-run.sh
#!/bin/bash

podman run --rm -it \
  --privileged \
  -v $(pwd)/containers:/var/lib/containers \
  -v $(pwd)/run:/var/run/containers \
  -v $(pwd):/workspace \
  -w /workspace \
  -e STORAGE_DRIVER=vfs \
  -e STORAGE_ROOTDIR=/workspace/containers \
  -e STORAGE_RUNROOT=/workspace/run \
  buildah-container "$@"
EOF

# Make the script executable
chmod +x buildah-run.sh

echo "Setup complete. You can now use './buildah-run.sh' to run Buildah commands."
