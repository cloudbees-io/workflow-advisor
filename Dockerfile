# syntax=docker/dockerfile:1
# Stage 1: Obtain a shell using BusyBox
FROM busybox:1.37.0-uclibc AS shell_builder
 # Use a BusyBox image with uclibc for static compilation
RUN chmod +x /bin/sh

# Stage 2 - final image
FROM gcr.io/distroless/static:nonroot

# Copy the 'sh' binary from BusyBox
COPY --from=shell_builder /bin/sh /bin/sh
COPY advisor /usr/local/bin/
