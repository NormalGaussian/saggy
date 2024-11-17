# Saggy

An ease of use tool for secret management in version control.

For use directly by those working with a codebase, and scripted in places such as CI/CD. 

Saggy is currently version `v0.8.0` - prior to v1, every release may break any previous interface.

## Quick Usage

Use an encrypted talos config with talosctl; the encrypted file is decrypted only for the duration of the command.

```bash
# Use a sops encrypted file or folder
# saggy with <file|folder> -- <command> [args...]
saggy with talosconfig.sops -- talosctl --config {} get rd
```

## Full usage example

```bash
# Generate a key
saggy keygen
#> A key is created in ./secrets/age.key
#> Its public key is added to ./secrets/public-age-keys.json

# Encrypt a secret
saggy encrypt mysecret
#> The encrypted version of the secret is stored in mysecret.sops, decryptable by all of the private keys matching public keys in ./secrets/public-age-keys.json

# Use a command (secureservice list ) with the secret
saggy with mysecret.sops -- secureservice --key {} list
```

## Installation instructions

// TODO: Use CI to generate a binary
// TODO: publish binaries to releases

## Quick Reference

```sh
# with
saggy with <location> -- <command> [args...]
# every '{}' present in command/args will be substituted with a decrypted version of `location`.

# keygen
saggy keygen

# encrypt
saggy encrypt <location> [destination]
# By default the destination is location sans extension + .sops + extension

# decrypt
saggy decrypt <location> [destination]

```

## Whats in a name?

Saggy comes from a poor quality portmanteau of [sops](https://getsops.io/) and [age](https://github.com/FiloSottile/age) (using what is understood to be the authors pronunciation), the two tools that the initial versions of saggy glue together.

## Missing features

These are features which you may be surprised / encounter issues with not being present:

* Single binary usage
    - Saggy currently relies on age and sops being installed on the host system; this is not ideal as configuration is unecessarily difficult!
* Support keys with passphrases
    - Saggy doesn't currently support asking for a passphrase to decrypt a key. This is wholly untested.
* SSH key encryption
    - This is a feature of age
* I *think* sops uses more complex logic to determine how to name the output file; and saggy likely does not match this.

## The path forwards

* Offer bundled age/sops, and default to it
* Support non-age encryption that sops supports
* Support more and better locations for keyfiles
* Support piping
* Officially support windows and darwin
* Support groups
* Support keys with passphrases
* SSH key encryption / decryption via age
* More conrete testing, including a more appropriate test runner
* Introduce `saggy env <varname>=<secret path>... -- <command> [args...]`
* Introduce env args to filepaths `saggy with -e <varname>=<secret>`
* Introduce `saggy gen-with-script <name> <secret path> -- <command> [args...]`
    - Generate a shell script that acts as a passthrough invocation for a command, ensuring it is always executed with a secret provided. e.g. `saggy gen-with-script cloosterctl ./talosconfig.sops -- talosctl --config {} @`, where `{}` specifies the file to decrypt, and `@` specifies what to do with trailing args (default is to append).
* Allow specifying shell
    * `--shell`, default is whatever the system has as `sh`
* Allow specifying substitution string
    * e.g. `-I` for xargs
* Allow for key rotation `saggy rotate [encrypted...]`

## The path already trodden

* Convert to go

## License

// TODO: add license information

## Contributing

// TODO: add contribution guidelines
