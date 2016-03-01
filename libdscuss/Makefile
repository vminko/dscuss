SRC = $(wildcard *.c)
OBJS = $(SRC:%.c=%.o)
LIB = libdscuss.a

.PHONY: all clean

$(LIB): core.o peer.o handshake.o network.o connection.o util.o config.o \
        entity.o user.o message.o \
        header.o packet.o payload_hello.o payload_advertisement.o \
        crypto.o crypto_hash.o crypto_pow.o\
        db.o topic.o subscriptions.o
	ar rcs $@ $^

-include $(SRC:%.c=%.d)

%.o: %.c
	$(CC) $(CFLAGS) -MMD -c $<

all: $(LIB)

default: $(LIB)

clean:
	rm -f $(OBJS) $(LIB) *.d
