<!--
SPDX-FileCopyrightText: 2025 SUSE LLC

SPDX-License-Identifier: Apache-2.0
-->

[![REUSE status](https://api.reuse.software/badge/git.fsfe.org/reuse/api)](https://api.reuse.software/info/git.fsfe.org/reuse/api)

# happy-compta.fr tools

This project is a Go library to interact with happy-compta.fr form code.
It is mimicking the user interactions with the website and is thus very tied to the website changes.

This project has been initiated as part of [SUSE Hackweek 25](https://hackweek.opensuse.org/projects/create-a-go-module-to-wrap-happy-compta-dot-fr).

Implemented features:
- List of the employees, providers, categories, bank accounts, accounting periods
- Creation of entries

A set of tools comes with the library to demonstrate its use.
- dumper: mostly meant for debugging, it dumps all the lists that can already be retrieved
- loader: adds entries from a CSV file and an optional folder of receipts
- csv-to-sepa: convert a CSV file into a SEPA transfer XML ([PAIN 001.001.03](https://www.cfonb.org/instruments-de-paiement/virement)) file
