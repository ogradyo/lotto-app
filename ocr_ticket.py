#!/usr/bin/env python3
import argparse
import json
import sys
from datetime import datetime

import cv2
import pytesseract
import numpy as np

def parse_args():
    ap = argparse.ArgumentParser()
    ap.add_argument("--image", required=True, help="path to ticket image")
    ap.add_argument("--debug", action="store_true", help="print debug info to stderr")
    return ap.parse_args()


# hard‑code TN Powerball layout once you measure a good reference photo:
# All coordinates relative to a normalized image size (e.g., 1200x2000)
NORMALIZED_W = 1200
NORMALIZED_H = 2000

DRAW_DATE_ROI = (100, 200, 1000, 260)  # x1, y1, x2, y2  (placeholder)

#PLAY_ROWS = [
#    (100, 400, 1100, 470),  # Play A
#    (100, 480, 1100, 550),  # Play B
#    (100, 560, 1100, 630),  # Play C
#    (100, 640, 1100, 710),  # Play D
#    (100, 720, 1100, 790),  # Play E
#]

PLAY_ROWS = [
    (160, 460, 1050, 520),  # A
    (160, 545, 1050, 605),  # B
    (160, 630, 1050, 690),  # C
    (160, 715, 1050, 775),  # D
    (160, 800, 1050, 860),  # E
]


# load and normalize the image
def load_and_normalize(path):
    img = cv2.imread(path)
    if img is None:
        raise RuntimeError(f"could not read image: {path}")

    # normalize size so ROIs are consistent
    img = cv2.resize(img, (NORMALIZED_W, NORMALIZED_H))
    return img

#Generic ROI OCR Helpers
# Digit only OCR for plays ---digits and flexible OCR for date ---text
def ocr_roi_digits(image, roi, debug=False):
    x1, y1, x2, y2 = roi
    crop = image[y1:y2, x1:x2]
    gray = cv2.cvtColor(crop, cv2.COLOR_BGR2GRAY)
    gray = cv2.GaussianBlur(gray, (3, 3), 0)
    _, thresh = cv2.threshold(gray, 0, 255,
                              cv2.THRESH_BINARY + cv2.THRESH_OTSU)

    config = "--oem 3 --psm 7 -c tessedit_char_whitelist=0123456789"
    text = pytesseract.image_to_string(thresh, config=config)
    if debug:
        print(f"[DEBUG] ROI {roi} -> '{text.strip()}'", file=sys.stderr)
    return text

def ocr_roi_text(image, roi, debug=False):
    x1, y1, x2, y2 = roi
    crop = image[y1:y2, x1:x2]
    gray = cv2.cvtColor(crop, cv2.COLOR_BGR2GRAY)
    gray = cv2.GaussianBlur(gray, (3, 3), 0)
    _, thresh = cv2.threshold(gray, 0, 255,
                              cv2.THRESH_BINARY + cv2.THRESH_OTSU)

    # allow digits and slashes for date like 12/31/2025
    config = "--oem 3 --psm 7 -c tessedit_char_whitelist=0123456789/"
    text = pytesseract.image_to_string(thresh, config=config)
    if debug:
        print(f"[DEBUG] DATE ROI {roi} -> '{text.strip()}'", file=sys.stderr)
    return text

# then we can parse the date and plays

def parse_draw_date(raw: str):
    raw = raw.strip()
    # Expect things like "12/31/25" or "12/31/2025"
    parts = [p for p in raw.split("/") if p]
    if len(parts) != 3:
        raise ValueError(f"cannot parse date from '{raw}'")
    m, d, y = parts
    if len(y) == 2:
        y = "20" + y
    dt = datetime(int(y), int(m), int(d))
    return dt.strftime("%Y-%m-%d")

def parse_play_line(raw: str):
    # Extract ints in order
    tokens = []
    current = ""
    for ch in raw:
        if ch.isdigit():
            current += ch
        else:
            if current:
                tokens.append(int(current))
                current = ""
    if current:
        tokens.append(int(current))

    # Need at least 6 numbers: 5 whites + 1 powerball
    if len(tokens) < 6:
        return None

    whites = tokens[:5]
    special = tokens[5]
    return {
        "white": whites,
        "special": special,
    }

def draw_debug_rois(image, output_path):
    debug_img = image.copy()
    # date box in blue
    x1, y1, x2, y2 = DRAW_DATE_ROI
    cv2.rectangle(debug_img, (x1, y1), (x2, y2), (255, 0, 0), 2)

    # play boxes in green
    for roi in PLAY_ROWS:
        x1, y1, x2, y2 = roi
        cv2.rectangle(debug_img, (x1, y1), (x2, y2), (0, 255, 0), 2)

    cv2.imwrite("debug_rois.png", debug_img)

# the main OCR function
def ocr_ticket(path, debug=False):
    image = load_and_normalize(path)
    if debug:
        draw_debug_rois(image, "debug_rois.png")

    # 1) draw date
    date_text = ocr_roi_text(image, DRAW_DATE_ROI, debug=debug)
    draw_date = None
    try:
        draw_date = parse_draw_date(date_text)
    except Exception as e:
        if debug:
            print(f"[DEBUG] date parse failed: {e}", file=sys.stderr)

    # 2) plays
    plays = []
    for roi in PLAY_ROWS:
        text = ocr_roi_digits(image, roi, debug=debug)
        if not text.strip():
            continue
        play = parse_play_line(text)
        if play is None:
            if debug:
                print(f"[DEBUG] play parse failed for '{text.strip()}'", file=sys.stderr)
            continue
        plays.append(play)

    return {
        "game": "POWERBALL",
        "draw_date": draw_date,  # may be None if parsing failed
        "plays": plays,
    }

#and the CLI entrypoint
def main():
    args = parse_args()
    try:
        result = ocr_ticket(args.image, debug=args.debug)
    except Exception as e:
        # Print a small error JSON; Go can look at "error" key
        print(json.dumps({"error": str(e)}))
        sys.exit(1)

    print(json.dumps(result))

if __name__ == "__main__":
    main()

