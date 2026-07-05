use std::alloc::{alloc, dealloc, Layout};
use std::slice;

use jxl::api::{
    states, JxlDecoder, JxlDecoderOptions, JxlOutputBuffer, JxlPixelFormat, ProcessingResult,
};
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

macro_rules! complete {
    ($e:expr) => {
        match $e {
            Ok(ProcessingResult::Complete { result }) => result,
            _ => return std::ptr::null_mut(),
        }
    };
}

/// Decode a JXL image; fills info=[w, h, depth(8), count] and returns a malloc'd
/// RGBA8 buffer of `count` frames, or null when config_only is set or on error.
#[no_mangle]
pub extern "C" fn decode(in_ptr: *const u8, in_len: i32, config_only: i32, info: *mut u32) -> *mut u8 {
    let input = unsafe { slice::from_raw_parts(in_ptr, in_len as usize) };
    let mut inp: &[u8] = input;

    let dec = JxlDecoder::<states::Initialized>::new(JxlDecoderOptions::default());
    let mut dec = complete!(dec.process(&mut inp));

    let (w, h) = dec.basic_info().size;
    unsafe {
        *info.add(0) = w as u32;
        *info.add(1) = h as u32;
        *info.add(2) = 8;
        *info.add(3) = 1;
    }
    if config_only != 0 {
        return std::ptr::null_mut();
    }

    let extra = dec.basic_info().extra_channels.len();
    dec.set_pixel_format(JxlPixelFormat::rgba8(extra));

    let frame_size = w * h * 4;
    let mut frames: Vec<Vec<u8>> = Vec::new();

    loop {
        let mut buf = vec![0u8; frame_size];
        {
            let mut obufs = vec![JxlOutputBuffer::new(&mut buf, h, w * 4)];
            let dec_fi = complete!(dec.process(&mut inp));
            dec = complete!(dec_fi.process(&mut inp, &mut obufs));
        }
        frames.push(buf);
        if !dec.has_more_frames() {
            break;
        }
    }

    let count = frames.len();
    unsafe {
        *info.add(3) = count as u32;
    }

    let out = malloc(frame_size * count);
    if out.is_null() {
        return out;
    }
    for (i, f) in frames.iter().enumerate() {
        unsafe {
            std::ptr::copy_nonoverlapping(f.as_ptr(), out.add(i * frame_size), frame_size);
        }
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
