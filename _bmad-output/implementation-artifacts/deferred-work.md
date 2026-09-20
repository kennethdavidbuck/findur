- source_spec: `/home/kennethb/Documents/Development/findur/_bmad-output/implementation-artifacts/spec-render-walking-skeleton.md`
  summary: Confirm that Render allocates `findur-api-kdb.onrender.com` and that the static-site API rewrite reaches it; otherwise serve the Vite build from Go.
  evidence: The Blueprint reserves a globally allocated hostname that cannot be proven before creation. A failed allocation or rewrite would make the current same-origin topology unusable and would settle this deferred finding.
