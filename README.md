# Science Data Transfer Protocol (SDTP) Client

This repository contains a CLI tool for downloading data from an SDTP server.

It supports listing available files, downloading those files, verifying their checksum,
and acknowledging successful downloads back to the server.

## Authentication

Authentication is done using x509 client certificates, therefore you must provide a
valid client certificate and private key. The certificate must be signed by a CA trusted
by the SDTP sever to successfully authenticate (connection will fail otherwise).

You can check the validity of the certificate by running the tool with no sub-command 
specified. This will check the certificate expiration and provide some basic information.

You can also use the `check` command to make a test connection to the server to ensure a
successful connection is possible.

## References
- Project Repository,
  https://github.com/asips/sdtp-client
- Science Data Transfer Protocol (SDTP) Interface Control Document (ICD), 
  https://www.earthdata.nasa.gov/s3fs-public/2023-11/423-ICD-027_SDTP_ICD_Original.pdf

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
