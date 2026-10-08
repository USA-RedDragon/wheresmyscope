FROM scratch
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/wheresmyscope /
ENTRYPOINT ["/wheresmyscope"]
