#  bigdata-importer

**bigdata-importer** is a tool for parsing and converting SQL dump files into multiple database targets with schema accuracy and minimal loss.

---

###  Purpose

The primary goal of this project is to not only parse SQL dump files but to provide a robust and accurate conversion system that maps field types and schema structures to target databases with minimal error.

By using the provided configuration file (`config.yaml`), the system operates fully isolated within your own server and connects securely to the specified database. Without exposing any credentials or sensitive data, it:

-  Converts parsed schema into the destination database’s native types  
-  Automatically applies field-level metadata such as **primary keys**, **indexes**, and **foreign keys**  
-  *(Planned)* Imports actual data into the generated tables when available

---

This makes the system **safe**, **extensible**, and **production-ready** for:

- Internal migrations  
- Data audits  
- Platform transitions

### Recent Update


The latest update introduces full MySQL → PostgreSQL schema conversion support.
The generator now correctly maps MySQL data types to PostgreSQL equivalents and ensures that:

All CREATE TABLE statements are generated first

All foreign keys and indexes are written after all tables, in proper order

The worker now merges all tables into a single unified SQL output file

This guarantees valid PostgreSQL-compatible schema dumps without cross-table dependency errors.

>  **This project is actively being developed and improved.**
