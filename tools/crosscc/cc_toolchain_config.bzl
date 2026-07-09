"""Resolution-only stub CC toolchain for macOS→linux/amd64 cross builds.

When cross-building OCI images from an arm64 macOS host to the amd64
cluster nodes (`--platforms=@rules_go//go/toolchain:linux_amd64`), Bazel's
toolchain resolution enumerates the registered rust prost / crate_universe
toolchains, whose crate targets reference `@rules_cc//cc:current_cc_toolchain`.
Those crates are never actually compiled into the Go/JS images (Go is built
pure, no rust lands in the image action graph), so this toolchain only needs
to *resolve* — its tool paths point at `/usr/bin/false` and are never invoked.
Forcing `--@rules_go//go/config:pure` guarantees rules_go never tries to use
it for cgo.

The accompanying `toolchain()` is `exec_compatible_with = [macos]` so
linux/amd64 CI runners never select this stub as their real CC toolchain.
"""

load("@rules_cc//cc:cc_toolchain_config_lib.bzl", "tool_path")

_TOOLS = ["gcc", "ld", "ar", "cpp", "nm", "objdump", "strip"]

def _impl(ctx):
    return cc_common.create_cc_toolchain_config_info(
        ctx = ctx,
        toolchain_identifier = "stub-linux-x86_64",
        host_system_name = "local",
        target_system_name = "x86_64-unknown-linux-gnu",
        target_cpu = "k8",
        target_libc = "unknown",
        compiler = "clang",
        abi_version = "unknown",
        abi_libc_version = "unknown",
        tool_paths = [tool_path(name = t, path = "/usr/bin/false") for t in _TOOLS],
    )

cc_toolchain_config = rule(
    implementation = _impl,
    attrs = {},
    provides = [CcToolchainConfigInfo],
)
