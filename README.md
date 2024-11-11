![release](https://github.com/jossware/kustomize-gopass/actions/workflows/release.yml/badge.svg)

# kustomize-gopass

kustomize-gopass is an [exec-based KRM function](https://kubectl.docs.kubernetes.io/guides/extending_kustomize/exec_krm_functions/) [Kustomize plugin](https://kubectl.docs.kubernetes.io/guides/extending_kustomize/) that injects secrets from [gopass](https://www.gopass.pw/) into Kubernetes `Secret` resources. This allows you to work with Kubernetes Secret resources in your Kustomize base without directly including the sensitive secret values in your manifests.

Here's how it works:

In your Kustomize base, you can include your `Secret` resources, but instead of embedding the actual secret values, you set the secret keys to point to the paths of the secrets stored in your pre-configured gopass password manager. When you run `kustomize build`, the kustomize-gopass plugin executes. It reads any gopass paths specified in the secret keys and retrieves the corresponding secret values from your gopass repository. The plugin then injects the retrieved secret values into the `Secret` resource(s) rendered by Kustomize.

This approach allows you to manage your sensitive data in gopass, while still maintaining the convenience of defining your Kubernetes resources in Kustomize.

## Table of Contents

- [kustomize-gopass](#kustomize-gopass)
    - [Table of Contents](#table-of-contents)
    - [Installation](#installation)
        - [Download pre-compiled binary](#download-pre-compiled-binary)
        - [go install](#go-install)
    - [Usage](#usage)
        - [Example](#example)
    - [When to use this?](#when-to-use-this)
    - [Development](#development)
    - [Running Tests](#running-tests)
    - [Contributing](#contributing)
    - [License](#license)

## Installation

### Download pre-compiled binary

kustomize-gopass is available on Linux, Mac, and Windows <sup>1</sup>.

1. Visit the [releases](https://github.com/jossware/kustomize-gopass/releases) page of this repository.
2. Download the appropriate archive for your operating system and architecture.
3. Extract the archive
4. Move the binary to a location in your PATH

### go install

```sh
go install github.com/jossware/kustomize-gopass@latest
```

## Usage

If you want to include a `Secret` in a Kustomize base that retrieves values from gopass, you need to:

1. Annotate it with the `config.kubernetes.io/function` to tell Kustomize what function to run.

    ``` yaml
    metadata:
      annotations:
        config.kubernetes.io/function: |
          exec:
            path: kustomize-gopass
    ```

The above assumes that the `kustomize-gopass` binary is in your `PATH`. If not, you can modify the above to the absolute path to the `kustomize-gopass` binary on your system.

2. Configure any `data` or `stringData` fields to use values stored in gopass. You do this by setting the field to a value like `gopass:<path/to/secret/in/gopass>`. For example:

    ``` yaml
    data:
      password: gopass:dev/db/password
    ```

Next, you need to configure Kustomize to treat the manifest for the `Secret` above as a [generator](https://kubectl.docs.kubernetes.io/guides/extending_kustomize/#specification-in-kustomizationyaml).

``` yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
generators:
  - my-secrets.yaml
# ...
```

3. Build

In order to run Kustomize with function support, you need to use the `--enable-alpha-plugins` and `--enable-exec` flags.

``` shell
kustomize build --enable-exec --enable-alpha-plugins my-base
```

### Example

my-secret.yaml

``` yaml
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: my-secrets
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: kustomize-gopass
data:
  dbpw: gopass:dev/db/password
  apikey: gopass:dev/thirdparty/app/apikey
```

kustomization.yaml

``` yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
generators:
  - my-secrets.yaml
# ...
```

``` shell
$ kustomize build --enable-exec --enable-alpha-plugins .
---
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: my-secrets
data:
  dbpw: <actual base64-encoded secret value>
  apikey: <actual base64-encoded secret value>
...
```

## When to use this?

We built kustomize-gopass for existing gopass users who want to more easily manage Kubernetes secrets for local development, side projects, or in homelab scenarios. Keep in mind that generating Kubernetes secrets client-side does make it rather easy for plain text secrets to leak into your terminal output, CI/CD logs, or elsewhere. For production, business-critical systems, we would lean towards something like the [Secrets Store CSI Driver](https://secrets-store-csi-driver.sigs.k8s.io/) or [External Secrets Operator](https://external-secrets.io/latest/) or any of the varied secrets-management solutions available in the Kubernetes ecosystem. 

## Development

To build and run the project locally, clone the repository and run:

```sh
git clone https://github.com/yourusername/kustomize-gopass.git
cd kustomize-gopass
go build
./kustomize-gopass
```

## Running Tests

To run tests, use the following command:

```sh
go test ./...
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

---

<sup>1</sup> _note_: kustomize-gopass has not been tested extensively on Windows. Please file an issue if you run into any problems.
