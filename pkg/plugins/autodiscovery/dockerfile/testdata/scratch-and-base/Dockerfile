ARG updatecli_version=v0.25.0
FROM updatecli/updatecli:${updatecli_version} AS base
FROM base AS step_0
FROM base AS step_1
FROM base AS step_2
FROM scratch
COPY --from=step_0 / /
COPY --from=step_1 / /
COPY --from=step_2 / /
