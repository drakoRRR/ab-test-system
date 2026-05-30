# Entity Relationship Diagram

```mermaid
erDiagram
    organizations {
        uuid        id          PK
        text        name
        timestamptz created_at
    }

    users {
        uuid        id           PK
        text        firebase_uid UK
        uuid        org_id       FK
        text        email        UK
        text        name
        text        photo_url
        user_role   role
        timestamptz created_at
        timestamptz updated_at
    }

    projects {
        uuid        id          PK
        uuid        org_id      FK
        text        name
        text        description
        timestamptz created_at
        timestamptz updated_at
    }

    api_keys {
        uuid        id          PK
        uuid        project_id  FK
        text        name
        text        key_hash    UK
        text        prefix
        timestamptz created_at
        timestamptz revoked_at
    }

    flags {
        uuid        id          PK
        uuid        project_id  FK
        text        key
        text        name
        boolean     enabled
        jsonb       rules
        timestamptz created_at
        timestamptz updated_at
    }

    experiments {
        uuid        id               PK
        uuid        project_id       FK
        uuid        flag_id          FK
        text        key
        text        name
        text        description
        text        status
        float       traffic_percent
        jsonb       variants
        timestamptz created_at
        timestamptz updated_at
        timestamptz started_at
        timestamptz ended_at
    }

    events {
        uuid             id            PK
        uuid             project_id    FK
        text             user_id
        uuid             experiment_id FK
        uuid             variant_id
        text             type
        text             name
        double_precision value
        timestamptz      ts
    }

    organizations ||--o{ users       : "has"
    organizations ||--o{ projects    : "has"
    projects      ||--o{ api_keys    : "has"
    projects      ||--o{ flags       : "has"
    projects      ||--o{ experiments : "has"
    projects      ||--o{ events      : "records"
    flags         |o--o{ experiments : "linked to"
    experiments   ||--o{ events      : "generates"
```

## Notes

- `experiments.variants` — JSONB array of `{id, key, name, weight}` objects. `variant_id` in `events` references a variant by UUID stored within this array — there is no separate `variants` table.
- `events.user_id` — external identifier (string from the client application), not a FK to `users`.
- `api_keys.key_hash` — only the hash is stored; the raw key is returned once on creation and never persisted.
- `flags.rules` — JSONB array of targeting rules (e.g. `[{"type": "percentage", "value": 0.5}]`).
- `users.org_id` — `ON DELETE SET NULL`; a user can exist without an organization.
- All other FK relations use `ON DELETE CASCADE`.
