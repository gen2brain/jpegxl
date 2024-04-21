LIBJXL_VERSION = v0.10.2

LIBJXL_SRC = $(PWD)/libjxl
LIBJXL_BUILD = $(LIBJXL_SRC)/build

WASI_SDK_PATH = /opt/wasi-sdk
export CC = $(WASI_SDK_PATH)/bin/clang --sysroot=$(WASI_SDK_PATH)/share/wasi-sysroot
export CXX = $(WASI_SDK_PATH)/bin/clang --sysroot=$(WASI_SDK_PATH)/share/wasi-sysroot
export CXXFLAGS = -msimd128

CMAKE_TOOLCHAIN_FILE=$(WASI_SDK_PATH)/share/cmake/wasi-sdk.cmake

BIN := jxl.wasm

all: $(BIN)

$(LIBJXL_SRC):
	git clone -b $(LIBJXL_VERSION) --depth 1 --recursive --jobs `nproc` https://github.com/libjxl/libjxl
	echo "int main(int argc, char** argv) { return 0; }" > $(LIBJXL_SRC)/third_party/brotli/c/tools/brotli.c
	sed -i '/^set(THREADS\|^find_package(Threads/d' $(LIBJXL_SRC)/CMakeLists.txt
	sed -i '/^include(jxl_threads/d' $(LIBJXL_SRC)/lib/CMakeLists.txt
	sed -i '/Threads::Threads/d' $(LIBJXL_SRC)/lib/jxl.cmake
	sed -i '/Threads::Threads/d' $(LIBJXL_SRC)/lib/jpegli.cmake
	mkdir -p $(LIBJXL_BUILD)
	test -d $@

$(LIBJXL_BUILD)/lib/libjxl.a: $(LIBJXL_SRC)
	cd $(LIBJXL_BUILD); \
	cmake $(LIBJXL_SRC) \
		-DCMAKE_BUILD_TYPE=Release \
		-DBUILD_SHARED_LIBS=0 \
		-DBUILD_TESTING=0 \
		-DJPEGXL_ENABLE_TOOLS=0 \
		-DJPEGXL_ENABLE_JPEGLI=0 \
		-DJPEGXL_ENABLE_JPEGLI_LIBJPEG=0 \
		-DJPEGXL_ENABLE_DOXYGEN=0 \
		-DJPEGXL_ENABLE_MANPAGES=0 \
		-DJPEGXL_ENABLE_BENCHMARK=0 \
		-DJPEGXL_ENABLE_EXAMPLES=0 \
		-DJPEGXL_ENABLE_SJPEG=0 \
		-DJPEGXL_ENABLE_WASM_TRHEADS=0 \
		-DJPEGXL_ENABLE_JNI=0 \
		-DJPEGXL_ENABLE_TRANSCODE_JPEG=0 \
		-DJPEGXL_ENABLE_BOXES=0 \
		-DCMAKE_TOOLCHAIN_FILE=$(CMAKE_TOOLCHAIN_FILE)

	cd $(LIBJXL_BUILD); \
	$(MAKE) -j$(shell nproc)

$(BIN): $(LIBJXL_BUILD)/lib/libjxl.a
	$(CC) \
		-O3 \
		-Wl,--no-entry \
		-Wl,--export=malloc \
		-Wl,--export=free \
		-Wl,--export=decode \
		-Wl,--export=encode \
		-mexec-model=reactor \
		-fno-exceptions \
		-mnontrapping-fptoint \
		-I ${LIBJXL_BUILD}/lib/include \
		-o $@ \
		-Wall \
		jxl.c \
		${LIBJXL_BUILD}/lib/libjxl.a \
		${LIBJXL_BUILD}/lib/libjxl_cms.a \
		${LIBJXL_BUILD}/third_party/highway/libhwy.a \
		${LIBJXL_BUILD}/third_party/brotli/libbrotlicommon.a \
		${LIBJXL_BUILD}/third_party/brotli/libbrotlienc.a \
		-lstdc++

.PHONY: clean

clean:
	-rm -rf $(LIBJXL_SRC)
