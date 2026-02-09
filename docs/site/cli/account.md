# account

Commands for viewing account information such as storage quota.

## Subcommands

- [account quota](#account-quota) -- Show storage quota

---

## account quota

Display the current storage quota usage for the account.

```bash
fastmail-cli account quota
```

### Output

Text output:

```
Storage: 2.0 GB / 50.0 GB (4.0%)
```

JSON output:

```json
{
  "used": 2147483648,
  "limit": 53687091200,
  "used_percent": 4.0
}
```

## See Also

- [CLI Reference](index.md)
