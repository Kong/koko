# Plugins schema from Kong

Plugins not included:
- oauth2

Plugins included with changes:
- Pre-function: schema copied from _schema.lua
- Post-function: schema copied from _schema.lua
- acme: removed check for shared_dict
- http-log: use patched.url instead of socket.url
