# Science Data Transfer Protocol (SDTP) Client

This repository contains a CLI tool for downloading data from an SDTP server.

It supports listing available files, downloading those files, verifying their checksum,
and acknowledging successful downloads back to the server.

## Authentication

Authentication is done using x509 client certificates, therefore you must provide a
valid client certificate and private key. The certificate must be signed by a CA trusted
by the SDTP sever to successfully authenticate (connection will fail otherwise).


## Verifying the Certificate

The `check` command can be used to verify the certificate expiration and also ensure 
the certificate can successfully connect to the server and request files.

If successful, you should see output similar to the following:
```
./sdtp-client check --cert path/to/cert.pem --key path/to/key.pem 
Certificate Ok!

    DN:              <Cert DN>
    Expiration Date: <Cert Expiration Date (a.k.a., NotAfter)>
    Days Left:       <Days until expiration>
    Issuer:          <Cert Issuer DN>

Successfully connected to server and performed a HEAD request to the /files endpoint.
```

If the certificate is expired it lines will be prefixed with `ERROR:` and the command
command will exit with status code 3.

If the certificate is valid but expires soon (see --check-cert-days) the lines will be 
preixed with `WARNING:` and the command will exit with status code 2.

All commands verify the certificate before attempting to connect to the server. This can
be disabled with the `--no-check-cert` flag.

When running the `check` command an actual request is made to the server to ensure the 
client can successfully connect and authenticate. If this fails the command will exit
with specific information about the failure, e.g., if the errors is due to authentication
or authorization.


## Listing Files

The `list` command can be used to list files availble on the server. This is a listing
only. No files are downloaded or acknowledged.


## Downloading Files

The `ingest` command can be used to download files from the server. The command will
perform a listing of all files available matching the provided tags and will download,
verify the checksum, and acknowledge each file.

Files as acknowledged by default, but this can be disabled with the `--no-ack` flag.


## References
- Project Repository,
  https://github.com/asips/sdtp-client
- Science Data Transfer Protocol (SDTP) Interface Control Document (ICD), 
  https://www.earthdata.nasa.gov/s3fs-public/2023-11/423-ICD-027_SDTP_ICD_Original.pdf

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
