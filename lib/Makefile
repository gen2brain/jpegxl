SHIM = target/wasm32-unknown-unknown/release/jpegxl_shim.wasm

# Build the Rust shim (jxl-oxide + zune-jpegxl) to scalar wasm, transpile to Go.
wasm2go: ../libjpegxl.go

$(SHIM): lib.rs Cargo.toml
	cargo build --target wasm32-unknown-unknown --release

# The sed lowercases the exported identifiers; it skips the data constants since
# the embedded Rust strings contain words like "New".
../libjpegxl.go: $(SHIM)
	wasm2go -pkg jpegxl -unsafe -tags wasm2go -o $@ $(SHIM)
	sed -i -E '/^const data[0-9]/! { s/\bModule\b/module/g; s/\bMemory\b/memory/g; s/\bNew\b/newModuleRaw/g; }' $@

.PHONY: wasm2go clean

clean:
	cargo clean
