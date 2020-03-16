# ChangeLog

## 2.0.0 (2020-03-16)
- make sure issued certificates never outlive the CA (#6)
- allow specifying `CA` as a value for `-not-after` to explicitly let the 
  certificate expire at the same time as the CA (#7)
- server and client certificates now expire after 1 year by default (instead of 
  together with the CA and 90 days respectively)

## 1.0.0 (2019-11-18)
- initial release
