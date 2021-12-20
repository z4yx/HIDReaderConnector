pcsclib.a: pcsclib.go reader.go
	go build -buildmode=c-archive -o $@ $^

hid_reader.so: ifdhandler.c pcsclib.a
	gcc $^ -I/usr/include/PCSC -o $@ -pthread -shared -fPIC

install: hid_reader.so hid-reader.conf
	cp hid_reader.so /usr/local/lib/
	cp hid-reader.conf /etc/reader.conf.d/
