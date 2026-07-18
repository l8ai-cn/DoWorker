# 3D Printing Rules

## Backend Selection

Choose the backend from the geometry:

| Geometry | Preferred backend |
| --- | --- |
| Bracket, jig, enclosure, holder, fitted part | JSCAD, build123d, CadQuery, or OpenSCAD |
| Organic shell, figurine, sculpted relief | Blender |
| Mixed precise and organic geometry | Create the dimension-critical body in CAD, then perform an explicit reviewed transfer |

Do not transfer between backends merely because the selected tool failed.

## FDM Starting Values

Use these only when the user has not provided printer-specific values:

| Feature | Starting value |
| --- | --- |
| Structural wall | 1.6-2.4 mm |
| Thin cosmetic wall | 1.2 mm |
| Sliding clearance | 0.30 mm per side |
| Snug removable fit | 0.20 mm per side |
| Press fit | Printer-specific calibration required |
| Minimum raised/engraved line | 0.8 mm |
| Minimum raised/engraved depth | 0.6 mm |
| Unsupported overhang | Keep at or below 45 degrees |
| Bridge | Keep under 10 mm unless printer capability is known |

Round mating values to practical nozzle and layer increments. Never claim a guaranteed fit without a calibration coupon or measured printer profile.

## Completion Evidence

The validation report must include:

- triangle count
- bounding-box dimensions in millimeters
- degenerate triangle count
- non-manifold edge count
- inconsistent edge orientation count
- enclosed volume
- pass/fail status

The README must identify assumptions that could change fit or strength.
