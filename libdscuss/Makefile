SRC = $(wildcard *.c)
OBJS = $(SRC:%.c=%.o)
LIB = libdscuss.a

.PHONY: all clean

$(LIB): api.o network.o connection.o util.o config.o entity.o message.o\
        packet.o packet_message.o peer.o crypto.o crypto_hash.o crypto_ecc.o crypto_pow.o
	ar rcs $@ $^

-include $(SRC:%.c=%.d)

%.o: %.c
	$(CC) $(CFLAGS) -MMD -c $<

all: $(LIB)

default: $(LIB)

clean:
	rm -f $(OBJS) $(LIB) *.d
