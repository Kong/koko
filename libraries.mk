

ATCROUTER_LIB = /tmp/lib/libatc_router.a

DEPS = $(ATCROUTER_LIB)

INSTALL_LIBS = libatc_router.so
INSTALLED_LIBS = $(addprefix /usr/lib/,$(INSTALL_LIBS))


$(ATCROUTER_LIB):
	./scripts/build-library.sh kong/go-atc-router make-lib.sh /tmp/lib

$(INSTALLED_LIBS): $(ATCROUTER_LIB)
	sudo -En ln -s /tmp/lib/$(INSTALL_LIBS) /usr/lib

