---
name: 3d-print-modeler
description: Create, revise, validate, and package dimensioned 3D-printable models. Use for functional parts, brackets, holders, enclosures, jigs, fixtures, decorative solids, or requests that must deliver STL, 3MF, STEP, OpenSCAD, JSCAD, CadQuery, build123d, or Blender files with printability checks.
---

# 3D Print Modeler

Build a real-scale solid, preserve editable source, and prove the exported mesh is printable before reporting completion.

## Workflow

1. Extract dimensions, loads, mating features, printer process, material, and required clearances. State low-risk assumptions in the report.
2. Choose one backend before modeling:
   - Use JSCAD, build123d, CadQuery, or OpenSCAD for functional and dimension-driven parts.
   - Use Blender for organic or artistic geometry only when Blender and its required control path are already available.
3. Record the selected backend and its availability check. If it is unavailable, stop with the exact missing dependency. Do not silently switch backends.
4. Model in millimeters with named parameters. Apply booleans before export and remove construction geometry.
5. Export the editable source and at least one mesh format.
6. Validate the final exported file, not only the source model.
7. Write the deliverables and concise printing guidance under `deliverables/`.

## Online JSCAD Backend

Use this backend for functional models in a Codex Node.js runtime:

```bash
cp -R <skill-directory>/assets/jscad-runtime /tmp/3d-print-jscad-runtime
npm ci --prefix /tmp/3d-print-jscad-runtime --ignore-scripts --no-audit --no-fund
```

The committed lockfile fixes transitive package versions and integrity hashes.
Treat installation failure as a blocked backend; do not switch modeling engines.
Keep the model script in `deliverables/<model-name>.jscad.mjs`. Serialize a single unioned `geom3` to binary or ASCII STL. Do not export overlapping unmerged solids.

## Printability Contract

- Use millimeters and report the final X/Y/Z dimensions.
- Produce a closed, manifold solid with nonzero volume and consistent triangle faces.
- Default FDM wall thickness to at least 1.6 mm unless the task provides a process-specific value.
- Default sliding clearance to 0.30 mm per side; use 0.20 mm per side for a snug fit only when justified.
- Apply explicit hole compensation when dimensional fit matters; record the nominal and modeled diameters.
- Add practical fillets or chamfers at load-bearing and bed-contact edges where the backend supports them.
- Avoid unsupported ceilings and overhangs above 45 degrees when the requested geometry permits.
- Keep embossed or engraved details at least 0.6 mm deep/high and 0.8 mm wide for ordinary FDM.

Read [print-rules.md](references/print-rules.md) when selecting detailed tolerances or preparing a slicer note.

## Validation

Run the bundled validator against every STL:

```bash
node <skill-directory>/scripts/validate-stl.mjs \
  deliverables/<model-name>.stl \
  deliverables/<model-name>-validation.json
```

Treat any nonzero exit as a failed model. Inspect and repair the geometry; do not weaken or bypass the validator.

For formats the validator does not parse, also export STL from the same final solid and validate that STL.

## Required Deliverables

- `deliverables/<model-name>.stl`
- `deliverables/<model-name>.<editable-source-extension>`
- `deliverables/<model-name>-validation.json`
- `deliverables/README.md` with dimensions, parameters, assumptions, material, layer/nozzle guidance, orientation, supports, clearances, and known limits

When requested and supported by the chosen backend, also export `3mf`, `step`, or `blend`.
