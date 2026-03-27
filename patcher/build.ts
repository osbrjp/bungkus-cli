await Bun.build({
	entrypoints: ["./src/index.ts"],
	outdir: "./dist",
	target: "node",
	format: "esm",
	external: ["fs", "path", "os", "crypto", "module", "url", "assert", "util", "events", "stream", "buffer"],
});
