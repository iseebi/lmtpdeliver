# lmtpdeliver

Simple LMTP deliver HTTP API

Your received eml files to deliver any Mail delivery agents. (ex. received in Amazon SES, deliver to Dovecot)

## Usage

### launch

```
$ lmtpdeliver
```

### Deliver a mail

```
$ curl -X POST \
       -F mail=@mail.eml \
       -F to=user@example.com \
       http://127.0.0.1:8080/delivery
```

