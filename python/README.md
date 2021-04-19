# Akeyless Custom Producers in Python

This folder includes Akeyless Custom Producers implemented in Python.

## Authentication

Please see the following example of credentials validation in custom producers
implemented in Python:

```
pip install requests
```

```python
import requests

akeyless_url = 'https://auth.akeyless.io/validate-producer-credentials'

# this token is received in AkeylessCreds header with every request
token = '<jwt token from AkeylessCreds header>'

# only the producer with the provided access ID and name will be able to access
# the webhook
access_id = '<expected-producer-access-id>'
item_name = '<expected-producer-name>'

body = {
        'creds': token,
        'expected_access_id': access_id,
        'expected_item_name': item_name
}
r = requests.post(akeyless_url, json=body)

if r.status_code != requests.codes.ok:
    raise Exception('code {}: {}'.format(r.status_code, r.json()))

producer_id = r.json()['access_id']
if producer_id != access_id:
    raise Exception('mismatched access id: {}'.format(producer_id))

print(r.json())
```
