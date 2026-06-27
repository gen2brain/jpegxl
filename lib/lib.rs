use std::alloc::{alloc, dealloc, Layout};
use std::io::Cursor;

use jxl_oxide::JxlImage;
use zune_core::bit_depth::BitDepth;
use zune_core::colorspace::ColorSpace;
use zune_core::options::EncoderOptions;
use zune_jpegxl::JxlSimpleEncoder;

const HDR: usize = 8;

/// Allocate `size` bytes, length-prefixed so `free` can rebuild the layout.
#[no_mangle]
pub extern "C" fn malloc(size: usize) -> *mut u8 {
    if size == 0 {
        return std::ptr::null_mut();
    }
    let total = size + HDR;
    let layout = Layout::from_size_align(total, HDR).unwrap();
    unsafe {
        let p = alloc(layout);
        if p.is_null() {
            return p;
        }
        (p as *mut usize).write(total);
        p.add(HDR)
    }
}

#[no_mangle]
pub extern "C" fn free(ptr: *mut u8) {
    if ptr.is_null() {
        return;
    }
    unsafe {
        let base = ptr.sub(HDR);
        let total = (base as *mut usize).read();
        dealloc(base, Layout::from_size_align(total, HDR).unwrap());
    }
}

/// Expand `ch`-channel 8-bit samples to interleaved RGBA8.
fn to_rgba8(src: &[u8], ch: usize, pixels: usize, dst: &mut [u8]) {
    for i in 0..pixels {
        let (r, g, b, a) = match ch {
            1 => (src[i], src[i], src[i], 255),
            2 => (src[i * 2], src[i * 2], src[i * 2], src[i * 2 + 1]),
            3 => (src[i * 3], src[i * 3 + 1], src[i * 3 + 2], 255),
            _ => (src[i * 4], src[i * 4 + 1], src[i * 4 + 2], src[i * 4 + 3]),
        };
        dst[i * 4] = r;
        dst[i * 4 + 1] = g;
        dst[i * 4 + 2] = b;
        dst[i * 4 + 3] = a;
    }
}

/// Decode a JXL image; fills info=[w, h, depth(8), count] and returns a malloc'd
/// RGBA8 buffer of `count` frames, or null when config_only is set or on error.
#[no_mangle]
pub extern "C" fn decode(in_ptr: *const u8, in_len: i32, config_only: i32, info: *mut u32) -> *mut u8 {
    let input = unsafe { std::slice::from_raw_parts(in_ptr, in_len as usize) };

    let image = match JxlImage::builder().read(Cursor::new(input)) {
        Ok(i) => i,
        Err(_) => return std::ptr::null_mut(),
    };

    let w = image.width();
    let h = image.height();
    let count = image.num_loaded_keyframes().max(1) as u32;
    unsafe {
        *info.add(0) = w;
        *info.add(1) = h;
        *info.add(2) = 8;
        *info.add(3) = count;
    }
    if config_only != 0 {
        return std::ptr::null_mut();
    }

    let frame_size = (w as usize) * (h as usize) * 4;
    let out = malloc(frame_size * count as usize);
    if out.is_null() {
        return out;
    }

    for i in 0..count as usize {
        let render = match image.render_frame(i) {
            Ok(r) => r,
            Err(_) => {
                free(out);
                return std::ptr::null_mut();
            }
        };
        let mut stream = render.stream();
        let ch = stream.channels() as usize;
        let pixels = stream.width() as usize * stream.height() as usize;
        let mut tmp = vec![0u8; pixels * ch];
        stream.write_to_buffer::<u8>(&mut tmp);

        let dst = unsafe { std::slice::from_raw_parts_mut(out.add(i * frame_size), frame_size) };
        to_rgba8(&tmp, ch, pixels, dst);
    }

    out
}

/// Losslessly encode RGBA8 pixels to JXL; returns a malloc'd buffer (length in
/// `size`) or null on failure.
#[no_mangle]
pub extern "C" fn encode(rgba: *const u8, width: i32, height: i32, size: *mut usize, _quality: i32, _effort: i32) -> *mut u8 {
    let w = width as usize;
    let h = height as usize;
    let input = unsafe { std::slice::from_raw_parts(rgba, w * h * 4) };

    let opts = EncoderOptions::new(w, h, ColorSpace::RGBA, BitDepth::Eight);
    let encoder = JxlSimpleEncoder::new(input, opts);

    let mut sink: Vec<u8> = Vec::new();
    if encoder.encode(&mut sink).is_err() {
        return std::ptr::null_mut();
    }

    let out = malloc(sink.len());
    if out.is_null() {
        return out;
    }
    unsafe {
        std::ptr::copy_nonoverlapping(sink.as_ptr(), out, sink.len());
        *size = sink.len();
    }
    out
}
