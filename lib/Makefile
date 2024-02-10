LIBJXL_VERSION = v0.9.2

LIBJXL_SRC = $(PWD)/libjxl
LIBJXL_BUILD = $(LIBJXL_SRC)/build

CMAKE_TOOLCHAIN_FILE = $(EMSDK)/upstream/emscripten/cmake/Modules/Platform/Emscripten.cmake

BIN := jxl.wasm

EMCMAKE := emcmake cmake
EMMAKE := emmake $(MAKE)
EMCC := emcc
EMCXX := em++

export CC = $(EMCC)
export CXX = $(EMCXX)
export CXXFLAGS = -msimd128 -no-pthread

all: jxl.wasm

$(LIBJXL_SRC):
	git clone -b $(LIBJXL_VERSION) --depth 1 --recursive --jobs `nproc` https://github.com/libjxl/libjxl
	mkdir -p $(LIBJXL_BUILD)
	test -d $@

$(LIBJXL_BUILD)/lib/libjxl.a: $(LIBJXL_SRC)
	cd $(LIBJXL_BUILD); \
	$(EMCMAKE) $(LIBJXL_SRC) \
		-DCMAKE_BUILD_TYPE=Release \
		-DBUILD_SHARED_LIBS=0 \
		-DBUILD_TESTING=0 \
		-DJPEGXL_ENABLE_TOOLS=0 \
		-DJPEGXL_ENABLE_JPEGLI=0 \
		-DJPEGXL_ENABLE_DOXYGEN=0 \
		-DJPEGXL_ENABLE_MANPAGES=0 \
		-DJPEGXL_ENABLE_BENCHMARK=0 \
		-DJPEGXL_ENABLE_EXAMPLES=0 \
		-DJPEGXL_BUNDLE_LIBPNG=0 \
		-DJPEGXL_ENABLE_JNI=0 \
		-DJPEGXL_ENABLE_SJPEG=0 \
		-DJPEGXL_ENABLE_WASM_TRHEADS=0 \
		-DJPEGXL_ENABLE_TRANSCODE_JPEG=0 \
		-DJPEGXL_ENABLE_BOXES=0 \
		-DCMAKE_TOOLCHAIN_FILE=$(CMAKE_TOOLCHAIN_FILE)

	cd $(LIBJXL_BUILD); \
	$(EMMAKE) -j$(shell nproc)

$(BIN): $(LIBJXL_BUILD)/lib/libjxl.a
	$(EMCC) \
		-O3 \
		--no-entry \
		-mnontrapping-fptoint \
		-sFILESYSTEM=0 \
		-sALLOW_MEMORY_GROWTH=1 \
		-sSTANDALONE_WASM=1 \
		-sUSE_PTHREADS=0 \
		-I ${LIBJXL_BUILD}/lib/include \
		-o $@ \
		-Wall \
		jxl.c \
		${LIBJXL_BUILD}/lib/libjxl.a \
		${LIBJXL_BUILD}/third_party/highway/libhwy.a

.PHONY: clean

clean:
	-rm -rf $(LIBJXL_SRC)
