all:
	g++ -fPIC -O3 -c schifra_reed_solomon.cpp -lstdc++ -lm
	g++ -dynamiclib -fPIC -o libschifra_reed_solomon.dylib schifra_reed_solomon.o
	go build uat.go
