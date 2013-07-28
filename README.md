# gologger
A TCP logger server writing in golang.


## Dependency
``` clojure
go1
```

## Build
``` clojure
cd /path/to/gologger && go build gologger
```

## Run
``` clojure
./gologger -c example.json
```

## Next generation message format

Maybe BSON is much better.

``` clojure

; a log memssage' format

; magic letter
; message version;
; key-value list. 
message := "GL" uint16 "\n" kv_list "\n"

kv_list := kv_element kv_list
         | nil

kv_element := "n" key "\n"				; NULL
     		| "i" key "\n" int32		; signed int32
      		| "I" key "\n" uint32		; unsigned int32
      		| "l" key "\n" int64		; signed int64
      		| "L" key "\n" uint64		; unsigned int64
      		| "b" key "\n" binary		; binary data
      		| "s" key "\n" string		; string in utf8 coding
      		| "d" key "\n" double		; 64-bit IEEE 754 floating point
      		| "t" key "\n"				; boolean true
      		| "f" key "\n"				; boolean false


binary := uint16 bytes

bytes := uint8 bytes
	   | nil

string := binary

key := ascii letters except "\n"


```