# Build the initial docker image for linuxboot (need to be executed one time)
cp Dockerfile.linux Dockerfile
docker build -t linuxboot .
# Execute the image
docker run --name linuxboot -v /tmp/volume:/volume linuxboot
# Get the result
cp /tmp/volume/linuxboot.bin .
# Destroy the past container
docker container rm linuxboot
