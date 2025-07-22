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

>  **This project is actively being developed and improved.**
