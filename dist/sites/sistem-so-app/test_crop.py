import cv2
import sys
import numpy as np

image_path = "public/storage/scans/1e0bmGRWaWC2SnsFjVAREUqBvDckSDwjZoxnu2QP.png"
image = cv2.imread(image_path)
if image is None:
    print("Cannot read image")
    sys.exit()
    
height, width = image.shape[:2]
gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
thresh = cv2.adaptiveThreshold(gray, 255, cv2.ADAPTIVE_THRESH_MEAN_C, cv2.THRESH_BINARY_INV, 15, 10)

# Find horizontal lines
horizontal_kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (40, 1))
detect_horizontal = cv2.morphologyEx(thresh, cv2.MORPH_OPEN, horizontal_kernel, iterations=2)
cnts = cv2.findContours(detect_horizontal, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
cnts = cnts[0] if len(cnts) == 2 else cnts[1]

# Collect all horizontal line y-coordinates
y_coords = []
for c in cnts:
    x, y, w, h = cv2.boundingRect(c)
    if w > width * 0.5: # It's a long horizontal line across the table
        y_coords.append(y)

# Sort y coordinates
y_coords = sorted(y_coords)
print("Found horizontal lines at Y:", y_coords)

# We expect 7 horizontal lines for a 5-item table (1 top header, 1 bottom header, 5 rows = 7 lines)
# Or just bounding boxes of rows
# Let's try bounding boxes instead
cnts_all = cv2.findContours(thresh, cv2.RETR_LIST, cv2.CHAIN_APPROX_SIMPLE)
cnts_all = cnts_all[0] if len(cnts_all) == 2 else cnts_all[1]

row_rects = []
for c in cnts_all:
    x, y, w, h = cv2.boundingRect(c)
    if w > width * 0.7 and h > 20 and h < height * 0.2:
        row_rects.append((x, y, w, h))

row_rects = sorted(row_rects, key=lambda r: r[1])
print("Found row rects:", len(row_rects))
for (x, y, w, h) in row_rects:
    cv2.rectangle(image, (x, y), (x+w, y+h), (0, 255, 0), 2)

cv2.imwrite("public/storage/scans/test_crop.jpg", image)
print("Saved to public/storage/scans/test_crop.jpg")
