# ChangeLog

## 2.0.0 (...)
- make sure issued certificates never outlive the CA (#6)
- allow specifiying `CA` as a value for `-not-after` to let the certificate
  expire at the same time as the CA (#7)
- server certificates now expire after 1 year by default (instead of together 
  with CA)

## 1.0.0 (2019-11-18)
- initial release
