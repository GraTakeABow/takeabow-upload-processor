docker run --rm \
   -v "$(pwd):/src" \
   -v /var/run/docker.sock:/var/run/docker.sock \
   xjewer/golang-builder:v1.2 \
   therealpenguin/takeabow:upload-processor

docker push therealpenguin/takeabow:upload-processor
