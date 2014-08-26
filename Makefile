CC=gcc
LD=gcc
CFLAGS=`pkg-config --cflags glib-2.0`
CFLAGS+=`pkg-config --cflags gio-2.0`
CFLAGS+=`pkg-config --cflags libcrypto`
CFLAGS+=-Wall -O2 -g
LDFLAGS=`pkg-config --libs glib-2.0`
LDFLAGS+=`pkg-config --libs gio-2.0`
LDFLAGS+=`pkg-config --libs libcrypto`
LIB_DIR=libdscuss
LIB=dscuss
OBJS=main.o
PROG=dscuss

.PHONY: all clean $(LIB_DIR) $(PROG)

$(PROG): $(LIB_DIR) $(OBJS)
	$(CC) $(LDFLAGS) $(OBJS) -L$(LIB_DIR) -l$(LIB) -o $(PROG)

%.o: %.c
	$(CC) $(CFLAGS) -c $<

$(LIB_DIR):
	$(MAKE) CC="$(CC)" LDFLAGS="$(LDFLAGS)" CFLAGS="$(CFLAGS)" -C $(LIB_DIR) $$target\
	  || exit 1;

all: $(PROG)

default: $(PROG)

clean:
	$(MAKE) -C $(LIB_DIR) clean
	rm -f $(OBJS) $(PROG)
