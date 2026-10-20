# Deploying the Gaia ADQL plugin to an existing Grafana

This guide covers building a distributable package of the plugin and installing it
on a Grafana instance you already run (bare metal / VM, Docker, or Kubernetes via
the Grafana Helm chart). For the local Docker Compose dev environment, see the
`Makefile` (`make start`).

| | |
|---|---|
| Plugin ID | `davidmillerpf9-gaia-adql-datasource` |
| Plugin type | Data source, with a Go backend (`gpx_gaia_adql`) |
| Grafana version | `>=10.0.0` declared; built and tested against **11.0.0** |
| Signed | No — Grafana must be configured to allow it (see [step 3](#3-allow-the-unsigned-plugin)) |
| Network | Grafana server needs outbound HTTPS to `gea.esac.esa.int:443` (or your TAP endpoint) |

## 1. Prerequisites (build machine)

- Go (the version in `go.mod` or newer)
- Node.js 20+ and npm
- `zip`

## 2. Build and package

Run from the repository root. Commands are bash/zsh syntax.

```bash
VERSION=$(node -p "require('./package.json').version")
PLUGIN_ID=davidmillerpf9-gaia-adql-datasource
STAGE=build/$PLUGIN_ID

# Frontend: compiles src/ into dist/ and stamps the version into plugin.json
npm ci
npm run build

# Stage the frontend output
rm -rf build && mkdir -p "$STAGE"
cp dist/module.js dist/plugin.json dist/CHANGELOG.md "$STAGE"/
cp -r dist/img "$STAGE"/ 2>/dev/null || true

# Backend: one binary per platform Grafana might run on.
# Grafana picks gpx_gaia_adql_<os>_<arch>[.exe] for its own platform,
# so trim this list to just what you need if you want a smaller zip.
for target in linux/amd64 linux/arm64 darwin/arm64 darwin/amd64 windows/amd64; do
  os=${target%/*}; arch=${target#*/}; ext=""
  [ "$os" = windows ] && ext=.exe
  CGO_ENABLED=0 GOOS=$os GOARCH=$arch \
    go build -ldflags "-w -s" -o "$STAGE/gpx_gaia_adql_${os}_${arch}${ext}" ./pkg
done

# Zip with the plugin ID as the top-level folder (Grafana requires this layout)
(cd build && zip -qr "$PLUGIN_ID-$VERSION.zip" "$PLUGIN_ID")
ls -l "build/$PLUGIN_ID-$VERSION.zip"
```

The result is `build/davidmillerpf9-gaia-adql-datasource-<version>.zip` with this layout:

```
davidmillerpf9-gaia-adql-datasource/
├── plugin.json
├── module.js
├── CHANGELOG.md
├── gpx_gaia_adql_linux_amd64
├── gpx_gaia_adql_linux_arm64
├── ...
```

The backend binaries must stay executable (`chmod +x`). `zip` preserves the mode bits,
but if you copy files by other means, check them.

> To release a new version, bump `version` in `package.json` (and add a
> `CHANGELOG.md` entry) before building — `plugin.json`'s `%VERSION%` is filled
> from it.

## 3. Allow the unsigned plugin

Grafana refuses to load unsigned plugins by default. Add the plugin ID to the
allow-list on the target Grafana — either in `grafana.ini`:

```ini
[plugins]
allow_loading_unsigned_plugins = davidmillerpf9-gaia-adql-datasource
```

or as an environment variable:

```bash
GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=davidmillerpf9-gaia-adql-datasource
```

The setting is a comma-separated list; if it already has entries, append to it
rather than replacing it. Grafana does **not** need to be in development mode.

If you'd rather not allow unsigned plugins, see [Signing](#optional-signing-the-plugin).

## 4. Install

Pick the method matching how your Grafana runs. All of them require a Grafana
restart afterwards.

### Option A — `grafana cli` (VM / bare metal)

```bash
sudo grafana cli \
  --pluginUrl /path/to/davidmillerpf9-gaia-adql-datasource-0.1.0.zip \
  plugins install davidmillerpf9-gaia-adql-datasource
sudo systemctl restart grafana-server
```

`--pluginUrl` accepts a local path or an `https://` URL. On older installs the
command is `grafana-cli` instead of `grafana cli`. If your Grafana uses a
non-default plugin directory, add `--pluginsDir <path>` (it must match
`[paths] plugins` in `grafana.ini`).

### Option B — copy the files manually

```bash
unzip davidmillerpf9-gaia-adql-datasource-0.1.0.zip -d /var/lib/grafana/plugins/
chown -R grafana:grafana /var/lib/grafana/plugins/davidmillerpf9-gaia-adql-datasource
sudo systemctl restart grafana-server
```

The plugin directory is `[paths] plugins` in `grafana.ini` (default
`/var/lib/grafana/plugins`). The filesystem must not be mounted `noexec`, since
Grafana executes the backend binary from it.

### Option C — Docker

Host the zip somewhere Grafana can download it from, then use the image's
`url;plugin-id` install syntax:

```bash
docker run -d -p 3000:3000 \
  -e GF_INSTALL_PLUGINS="https://example.com/davidmillerpf9-gaia-adql-datasource-0.1.0.zip;davidmillerpf9-gaia-adql-datasource" \
  -e GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=davidmillerpf9-gaia-adql-datasource \
  grafana/grafana:11.0.0
```

Alternatively, mount the unzipped folder into the container's plugin directory:

```bash
-v /path/to/davidmillerpf9-gaia-adql-datasource:/var/lib/grafana/plugins/davidmillerpf9-gaia-adql-datasource
```

> `GF_INSTALL_PLUGINS` is the Grafana 11.x mechanism. Grafana 12 deprecates it in
> favour of the `[plugins] preinstall` settings — check the docs for your version.

### Option D — Kubernetes (Grafana Helm chart)

In your `values.yaml`:

```yaml
plugins:
  - https://example.com/davidmillerpf9-gaia-adql-datasource-0.1.0.zip;davidmillerpf9-gaia-adql-datasource

grafana.ini:
  plugins:
    allow_loading_unsigned_plugins: davidmillerpf9-gaia-adql-datasource
```

Then `helm upgrade` your release. If you already have `plugins` or
`allow_loading_unsigned_plugins` set, merge these entries into the existing lists.

## 5. Verify the plugin loaded

In the Grafana server log you should see:

```
level=warn msg="Permitting unsigned plugin. This is not recommended" pluginId=davidmillerpf9-gaia-adql-datasource
level=info msg="Plugin registered" pluginId=davidmillerpf9-gaia-adql-datasource
```

Or via the API:

```bash
curl -s -u admin:<password> https://grafana.example.com/api/plugins/davidmillerpf9-gaia-adql-datasource/settings
```

If the log says the plugin is unsigned and was skipped, the allow-list from step 3
isn't in effect. If it reports it can't find or start the executable, the binary
for that server's OS/architecture is missing from the package or isn't executable.

## 6. Create the data source

**Through the UI:** *Connections → Data sources → Add data source → Gaia ADQL*.
Leave *TAP Service URL* empty to use the public ESA archive
(`https://gea.esac.esa.int/tap-server/tap`), then click **Save & test**. You should
see "TAP service is reachable at …".

**Through provisioning:** drop a file like this into
`<provisioning dir>/datasources/` and restart Grafana:

```yaml
apiVersion: 1

datasources:
  - name: Gaia ADQL
    type: davidmillerpf9-gaia-adql-datasource
    access: proxy
    jsonData:
      baseUrl: https://gea.esac.esa.int/tap-server/tap
```

**Through the API:**

```bash
curl -s -u admin:<password> -H 'Content-Type: application/json' \
  -X POST https://grafana.example.com/api/datasources \
  -d '{"name":"Gaia ADQL","type":"davidmillerpf9-gaia-adql-datasource","access":"proxy",
       "jsonData":{"baseUrl":"https://gea.esac.esa.int/tap-server/tap"}}'
```

Then try a query in Explore:

```sql
SELECT TOP 10 source_id, ra, dec, parallax, phot_g_mean_mag
FROM gaiadr3.gaia_source
WHERE parallax > 100
```

### Sample dashboard (optional)

`provisioning/dashboards/gaia-sample.json` can be imported via *Dashboards → New →
Import*. Two of its panels use the `ae3e-plotly-panel` plugin, which must be
installed separately (`grafana cli plugins install ae3e-plotly-panel`).

## Upgrading and removing

- **Upgrade:** build a new zip, install it the same way (`grafana cli` overwrites the
  existing directory; for manual installs, delete the old directory first), restart
  Grafana. Existing data sources and dashboards are kept.
- **Remove:** `grafana cli plugins remove davidmillerpf9-gaia-adql-datasource` (or
  delete the directory), then restart. Delete any Gaia ADQL data sources first.

## Optional: signing the plugin

A signed plugin loads without the allow-list. For a plugin that isn't published to
the Grafana catalog, use a **private** signature, which is tied to the URL(s) of the
Grafana instance(s) that will run it:

1. Create a Grafana Cloud account and an access policy token with the
   `plugins:write` scope (grafana.com → *My Account → Access Policies*).
2. After staging the files (step 2, before zipping), sign the staged directory:

   ```bash
   export GRAFANA_ACCESS_POLICY_TOKEN=<token>
   npx --yes @grafana/sign-plugin@latest \
     --distDir build/davidmillerpf9-gaia-adql-datasource \
     --rootUrls https://grafana.example.com/
   ```

   This writes `MANIFEST.txt` into the directory. Sign *after* all backend binaries
   are in place, since they're included in the signature, then zip as usual.
3. `--rootUrls` must match the target Grafana's `[server] root_url` exactly.
   Signing for a new URL means re-signing and repackaging.

## Known limitations

- **No HTTP proxy support.** The backend builds its own HTTP transport and ignores
  `HTTP_PROXY` / `HTTPS_PROXY`. If your Grafana needs a proxy for outbound traffic,
  the plugin won't reach the TAP service until this is fixed in
  `pkg/plugin/datasource.go`.
- **The API key field isn't used.** The config editor stores an API key, but the
  backend doesn't send it. Only unauthenticated TAP services work today.
- **No logo.** `plugin.json` refers to `img/logo.svg`, which doesn't exist yet, so the
  data source shows a broken image in the UI. It doesn't affect anything else.
- **Long queries.** Requests time out after 120 s and use TAP's synchronous endpoint.
  Keep queries bounded (`TOP n`, the *Max rows* setting) for large cone searches.
