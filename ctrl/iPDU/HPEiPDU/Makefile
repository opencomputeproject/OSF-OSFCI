libgpg:
	\rm -f libgpg-error-1.20.tar.gz
	wget https://gnupg.org/ftp/gcrypt/libgpg-error/libgpg-error-1.20.tar.gz
	gunzip libgpg-error-1.20.tar.gz
	tar xf libgpg-error-1.20.tar
	\rm -f libgpg-error-1.20.tar
	cd libgpg-error-1.20 ; ./configure --prefix=/usr/local/old ; make -j 4 install

libgcrypt: libgpg
	\rm -f libgcrypt-1.5.4.tar.gz
	wget https://gnupg.org/ftp/gcrypt/libgcrypt/libgcrypt-1.5.4.tar.gz
	gunzip libgcrypt-1.5.4.tar
	tar xf libgcrypt-1.5.4.tar
	\rm -f libgcrypt-1.5.4.tar
	export PATH=$$PATH:/usr/local/old/bin ; cd libgcrypt-1.5.4 ; ./configure --prefix=/usr/local/old ; make -j 4 install

gnutls_good: libgcrypt 
	\rm -f gnutls-2.10.5.tar.bz2
	wget https://ftp.gnu.org/gnu/gnutls/gnutls-2.10.5.tar.bz2
	bunzip2 gnutls-2.10.5.tar.bz2 
	tar xf gnutls-2.10.5.tar
	\rm -f gnutls-2.10.5.tar
	export PATH=$$PATH:/usr/local/old/bin ; cd gnutls-2.10.5 ; ./configure --prefix=/usr/local/old ; patch gl/stdio.in.h ../patches/stdio.patch ;  patch lib/gl/stdio.in.h ../patches/stdio.patch ; make -j 4 install
	touch gnutls_good

all: gnutls_good
	cc -I/usr/local/old/include -o iPDU_HPE client.c -L/usr/local/old/lib -lgnutls
	rm gnutls_good
install: all
	cp iPDU_HPE /usr/local/bin
	cp iPDU_Cmd /usr/local/bin
	cp ../iPDU_activate /usr/local/bin
	chmod -Rf 755 /usr/local/bin/iPDU_HPE /usr/local/bin/iPDU_Cmd /usr/local/bin/iPDU_activate
	/usr/local/bin/iPDU_activate /usr/local/bin/iPDU_Cmd

clean:
