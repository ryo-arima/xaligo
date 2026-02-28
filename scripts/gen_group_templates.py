#!/usr/bin/env python3
"""
Generate template files for each AWS Group type.

Output:
  etc/resources/aws/templates/excalidraw/  *.excalidraw
  etc/resources/aws/templates/xal/         *.xal

Run from the project root:
  python3 scripts/gen_group_templates.py
"""

import base64
import hashlib
import json
import os
import time

SVG_DIR = "etc/resources/aws/svg/Architecture-Group-Icons"
OUT_DIR_EXCALIDRAW = "etc/resources/aws/templates/excalidraw"
OUT_DIR_XAL        = "etc/resources/aws/templates/xal"

# (tag, label, stroke_color, stroke_style, stroke_width, icon_filename_or_None)
GROUPS = [
    ("aws-cloud",                    "AWS Cloud",                       "#000000", "solid",  2, "AWS-Cloud-logo_32.svg"),
    ("aws-cloud-alt",               "AWS Cloud",                       "#000000", "solid",  2, "AWS-Cloud_32.svg"),
    ("region",                      "Region",                          "#00A1C9", "dashed", 2, "Region_32.svg"),
    ("availability-zone",           "Availability Zone",               "#00A1C9", "dashed", 2, None),
    ("security-group",              "Security group",                  "#CC0000", "dashed", 2, None),
    ("auto-scaling-group",          "Auto Scaling group",              "#E7601B", "dashed", 2, "Auto-Scaling-group_32.svg"),
    ("vpc",                         "Virtual private cloud (VPC)",     "#8C4FFF", "solid",  2, "Virtual-private-cloud-VPC_32.svg"),
    ("private-subnet",              "Private subnet",                  "#00A1C9", "solid",  2, "Private-subnet_32.svg"),
    ("public-subnet",               "Public subnet",                   "#3F8624", "solid",  2, "Public-subnet_32.svg"),
    ("server-contents",             "Server contents",                 "#7A7C7F", "solid",  2, "Server-contents_32.svg"),
    ("corporate-data-center",       "Corporate data center",           "#7A7C7F", "solid",  2, "Corporate-data-center_32.svg"),
    ("ec2-instance-contents",       "EC2 instance contents",           "#E7601B", "solid",  2, "EC2-instance-contents_32.svg"),
    ("spot-fleet",                  "Spot Fleet",                      "#E7601B", "solid",  2, "Spot-Fleet_32.svg"),
    ("aws-account",                 "AWS account",                     "#E7008A", "solid",  2, "AWS-Account_32.svg"),
    ("aws-iot-greengrass-deployment","AWS IoT Greengrass Deployment",  "#3F8624", "solid",  2, "AWS-IoT-Greengrass-Deployment_32.svg"),
    ("aws-iot-greengrass",          "AWS IoT Greengrass",              "#3F8624", "solid",  2, None),
    ("elastic-beanstalk-container", "Elastic Beanstalk container",     "#E7601B", "solid",  2, None),
    ("aws-step-functions-workflow", "AWS Step Functions workflow",     "#E7008A", "solid",  2, None),
    ("generic-group",               "Generic group",                   "#AAB7B8", "dashed", 1, None),
]

W, H = 400, 280   # template canvas size
PADDING = 16
ICON_SIZE = 32
FONT_SIZE = 14
FONT_FAMILY = 2   # 1=Virgil(hand) 2=Helvetica(normal) 3=Cascadia(code)


def file_id(name: str) -> str:
    return hashlib.md5(name.encode()).hexdigest()[:16]


def svg_data_url(path: str) -> str:
    with open(path, "rb") as f:
        return "data:image/svg+xml;base64," + base64.b64encode(f.read()).decode()


def uid(seed: str) -> str:
    return hashlib.md5(seed.encode()).hexdigest()[:12]


def make_excalidraw(tag, label, stroke_color, stroke_style, stroke_width, icon_path):
    now_ms = int(time.time() * 1000)
    elements = []
    files = {}

    # ── outer border rectangle ──────────────────────────────────
    elements.append({
        "id": uid(f"{tag}-rect"),
        "type": "rectangle",
        "x": 0, "y": 0,
        "width": W, "height": H,
        "strokeColor": stroke_color,
        "backgroundColor": "transparent",
        "fillStyle": "solid",
        "strokeWidth": stroke_width,
        "strokeStyle": stroke_style,
        "roughness": 0,
        "opacity": 100,
        "angle": 0,
        "version": 1,
        "versionNonce": 1,
        "isDeleted": False,
        "groupIds": [],
        "frameId": None,
        "boundElements": None,
        "updated": now_ms,
        "link": None,
        "locked": False,
    })

    # icon top-left corner aligns with border corner at (0, 0)
    text_x = ICON_SIZE + 4
    text_y = (ICON_SIZE - FONT_SIZE) // 2  # vertically centered with icon

    # ── icon image ──────────────────────────────────────────────
    if icon_path and os.path.exists(icon_path):
        fid = file_id(os.path.basename(icon_path))
        data_url = svg_data_url(icon_path)
        elements.append({
            "id": uid(f"{tag}-icon"),
            "type": "image",
            "x": 0, "y": 0,  # corner-aligned with border
            "width": ICON_SIZE, "height": ICON_SIZE,
            "fileId": fid,
            "status": "saved",
            "scale": [1, 1],
            "strokeColor": "transparent",
            "backgroundColor": "transparent",
            "fillStyle": "solid",
            "strokeWidth": 1,
            "strokeStyle": "solid",
            "roughness": 0,
            "opacity": 100,
            "angle": 0,
            "version": 1,
            "versionNonce": 2,
            "isDeleted": False,
            "groupIds": [],
            "frameId": None,
            "boundElements": None,
            "updated": now_ms,
            "link": None,
            "locked": False,
        })
        files[fid] = {
            "mimeType": "image/svg+xml",
            "id": fid,
            "dataURL": data_url,
            "created": now_ms,
            "lastRetrieved": now_ms,
        }
    else:
        text_x = 4  # small left margin when no icon
        text_y = (ICON_SIZE - FONT_SIZE) // 2

    # ── label text ───────────────────────────────────────────────
    elements.append({
        "id": uid(f"{tag}-label"),
        "type": "text",
        "x": text_x, "y": text_y,
        "width": W - text_x - PADDING, "height": FONT_SIZE + 4,
        "text": label,
        "fontSize": FONT_SIZE,
        "fontFamily": FONT_FAMILY,
        "textAlign": "left",
        "verticalAlign": "middle",
        "strokeColor": stroke_color,
        "backgroundColor": "transparent",
        "fillStyle": "solid",
        "strokeWidth": 1,
        "strokeStyle": "solid",
        "roughness": 0,
        "opacity": 100,
        "angle": 0,
        "version": 1,
        "versionNonce": 3,
        "isDeleted": False,
        "groupIds": [],
        "frameId": None,
        "boundElements": None,
        "updated": now_ms,
        "link": None,
        "locked": False,
        "containerId": None,
        "originalText": label,
        "lineHeight": 1.25,
    })

    # ── sample child placeholder ─────────────────────────────────
    # starts just below the icon/label row
    cy = ICON_SIZE + 8
    elements.append({
        "id": uid(f"{tag}-child"),
        "type": "rectangle",
        "x": PADDING, "y": cy,
        "width": W - PADDING * 2, "height": H - cy - PADDING,
        "strokeColor": "#999999",
        "backgroundColor": "#f5f5f5",
        "fillStyle": "solid",
        "strokeWidth": 1,
        "strokeStyle": "dashed",
        "roughness": 0,
        "opacity": 60,
        "angle": 0,
        "version": 1,
        "versionNonce": 4,
        "isDeleted": False,
        "groupIds": [],
        "frameId": None,
        "boundElements": None,
        "updated": now_ms,
        "link": None,
        "locked": False,
    })
    elements.append({
        "id": uid(f"{tag}-child-label"),
        "type": "text",
        "x": PADDING + 8,
        "y": cy + (H - cy - PADDING) // 2 - 8,
        "width": W - PADDING * 2 - 16, "height": 20,
        "text": "( place resources here )",
        "fontSize": 12,
        "fontFamily": FONT_FAMILY,
        "textAlign": "center",
        "verticalAlign": "middle",
        "strokeColor": "#aaaaaa",
        "backgroundColor": "transparent",
        "fillStyle": "solid",
        "strokeWidth": 1,
        "strokeStyle": "solid",
        "roughness": 0,
        "opacity": 100,
        "angle": 0,
        "version": 1,
        "versionNonce": 5,
        "isDeleted": False,
        "groupIds": [],
        "frameId": None,
        "boundElements": None,
        "updated": now_ms,
        "link": None,
        "locked": False,
        "containerId": None,
        "originalText": "( place resources here )",
        "lineHeight": 1.25,
    })

    return {
        "type": "excalidraw",
        "version": 2,
        "source": "xaligo",
        "elements": elements,
        "appState": {
            "viewBackgroundColor": "#ffffff",
            "gridSize": None,
        },
        "files": files,
    }


def make_xal(tag: str, label: str, icon_file) -> str:
    icon_comment = f"icon: {icon_file}" if icon_file else "no icon"
    return f"""<!-- {tag}: {label} ({icon_comment}) -->
<frame width="600" height="440" class="pa-4">
  <{tag} title="{label}">
    <row gap="16">
      <col span="6" class="pa-2">
        <card title="Resource A" />
      </col>
      <col span="6" class="pa-2">
        <card title="Resource B" />
      </col>
    </row>
  </{tag}>
</frame>
"""


def main():
    project_root = os.getcwd()
    svg_dir_abs = os.path.join(project_root, SVG_DIR)
    out_excalidraw = os.path.join(project_root, OUT_DIR_EXCALIDRAW)
    out_xal = os.path.join(project_root, OUT_DIR_XAL)
    os.makedirs(out_excalidraw, exist_ok=True)
    os.makedirs(out_xal, exist_ok=True)

    for tag, label, stroke_color, stroke_style, stroke_width, icon_file in GROUPS:
        icon_path = None
        if icon_file:
            icon_path = os.path.join(svg_dir_abs, icon_file)

        scene = make_excalidraw(tag, label, stroke_color, stroke_style, stroke_width, icon_path)
        excalidraw_path = os.path.join(out_excalidraw, f"{tag}.excalidraw")
        with open(excalidraw_path, "w", encoding="utf-8") as f:
            json.dump(scene, f, ensure_ascii=False, indent=2)

        xal_path = os.path.join(out_xal, f"{tag}.xal")
        with open(xal_path, "w", encoding="utf-8") as f:
            f.write(make_xal(tag, label, icon_file))

        print(f"  {tag}")

    print(f"\nDone. {len(GROUPS)} templates generated.")
    print(f"  excalidraw -> {OUT_DIR_EXCALIDRAW}/")
    print(f"  xal        -> {OUT_DIR_XAL}/")


if __name__ == "__main__":
    main()
