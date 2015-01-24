CC=gcc
LD=gcc
CFLAGS+=`pkg-config --cflags glib-2.0`
CFLAGS+=`pkg-config --cflags gio-2.0`
CFLAGS+=`pkg-config --cflags libcrypto`
CFLAGS+=-Wall -O2 -g
LDFLAGS=`pkg-config --libs glib-2.0`
LDFLAGS+=`pkg-config --libs gio-2.0`
LDFLAGS+=`pkg-config --libs libcrypto`
LDFLAGS+=`pkg-config --libs sqlite3`
LIB_DIR=libdscuss
LIB=dscuss
OBJS=main.o

.PHONY: all clean $(LIB_DIR) dscuss kdf_bench

dscuss: $(LIB_DIR) $(OBJS)
	$(CC) $(LDFLAGS) $(OBJS) -L$(LIB_DIR) -l$(LIB) -o dscuss

kdf_bench: kdf_bench.o
	$(CC) $(LDFLAGS) kdf_bench.o -L$(LIB_DIR) -l$(LIB) -o kdf_bench

%.o: %.c
	$(CC) $(CFLAGS) -c $<

$(LIB_DIR):
	$(MAKE) CC="$(CC)" LDFLAGS="$(LDFLAGS)" CFLAGS="$(CFLAGS)" -C $(LIB_DIR) $$target\
	  || exit 1;

all: dscuss kdf_bench

default: dscuss

clean:
	$(MAKE) -C $(LIB_DIR) clean
	rm -f $(OBJS) dscuss  kdf_bench.o kdf_bench
