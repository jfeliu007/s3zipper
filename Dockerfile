FROM golang:1.10
WORKDIR /app
# Set an env var that matches your github repo name, replace treeder/dockergo here with your repo name
ENV SRC_DIR=/go/src/github.com/creativedrive/stream-download

# Add the source code:
COPY ./vendor/ $SRC_DIR/vendor
COPY ./. $SRC_DIR/

# run entrypoint.sh:
RUN cd $SRC_DIR; cp entry.sh /app/entry.sh; chmod 777 /app/entry.sh
CMD "./entry.sh"