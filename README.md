# mMailer
**Unify email services into 1 api for transactional email, have redundancies, traffic and avoid vendor locking**

```bash
SELECT_STRATEGY=RoundRobin \
SERVICES="generic:smtp://user:pass@smtp.server.com:25 mailjet:pubkeyXXXX:secretkeyYYYY" \
mmailerd
 
# Services:
#  - Generic: posthooks are not implmented, adding smtp://user:pass@smtp.server.com:25
#  - Mailjet: add the following posthook url  example.com/path/to/mmailer/posthook?service=mailjet
# Select Strategy: RoundRobin
# Retry Strategy:  None

# > Send mail by HTTP POST example.com/path/to/mmailer/send?key=

# Starting server, :8080
```


```bash
curl 'http://localhost:8080/send' \
  --data-binary \
  $'{"from": {"email": "jon.doe@example.com",
              "name": "Jon Doe"    },
     "to": [{"email": "jane.doe@example.com",
             "name": "Jane Doe"}],
     "subject": "A subject",
     "html": "<p><strong>Hey</strong> this is where the html gose</p>"
    }' --compressed
```