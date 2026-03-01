---
applyTo: "**"
---

# xaligo — Diagram Creation Guide

Standard workflow for creating Excalidraw diagrams.

---

## Step 1 — Find Service IDs

`etc/resources/aws/service-index.csv` maps service IDs to service names.
Use `grep` to search for the services you need.

```bash
# Format: id,service
grep -i "ec2"          etc/resources/aws/service-index.csv
grep -i "rds\|aurora"  etc/resources/aws/service-index.csv
grep -i "cloudfront"   etc/resources/aws/service-index.csv
```

Example output:
```
27,Amazon EC2
117,Amazon RDS
1178,Amazon CloudFront
```

---

## Step 2 — Create services.csv

`services.csv` lists the services to include in the diagram.

**Format:** `id,OfficialName,Abbreviation,Summary,Usage,Notes`

- Column 1 (`id`) as a number → icon is fetched from service-catalog.csv.
- Lines starting with `#` are treated as comments and ignored.
- `Abbreviation`, when set, is used as the icon label.

```csv
# 3-tier Web Architecture service list
# Format: id,OfficialName,Abbreviation,Summary,Usage,Notes
1179,Amazon Route 53,R53,DNS web service,Domain name resolution and health checks,
1178,Amazon CloudFront,CF,Content Delivery Network (CDN),Fast delivery of static/dynamic content,
1182,Elastic Load Balancing,ELB,Load balancing service,Distribute traffic across multiple instances,
27,Amazon EC2,EC2,Virtual server,Application tier,
1020,Amazon Simple Storage Service,S3,Object storage,Static assets and log storage,
117,Amazon RDS,RDS,Relational database,Data persistence,
```

Reference: [examples/services.csv](../../examples/services.csv)

---

## Step 3 — Create a .xal file

Use `<item id="N" />` to place service icons in the layout.
`N` is the service ID from the first column of service-index.csv.

```xml
<frame width="1440" height="900" class="pa-4" item-size="48">
  <text>3-tier Web Architecture</text>

  <row gap="16">
    <col span="4">
      <card title="Tier 1 — Presentation" />
      <item id="1179" />  <!-- Route 53 -->
      <item id="1178" />  <!-- CloudFront -->
    </col>

    <col span="4">
      <card title="Tier 2 — Application" />
      <item id="1182" />  <!-- ELB -->
      <item id="27"   />  <!-- EC2 -->
    </col>

    <col span="4">
      <card title="Tier 3 — Data" />
      <item id="117"  />  <!-- RDS -->
      <item id="113"  />  <!-- ElastiCache -->
    </col>
  </row>
</frame>
```

Reference: [examples/sample.xal](../../examples/sample.xal)
DSL specification: [xal-spec.instructions.md](xal-spec.instructions.md)

---

## Step 4 — Generate the Excalidraw file

```bash
xaligo generate excalidraw \
  --xal examples/sample.xal \
  -o output/sample.excalidraw \
  --services examples/services.csv
```

`--services` is a **required parameter**. The services listed in services.csv are
added as a legend on the right side of the frame.

> **Note:** Create the output directory if it does not already exist.
> ```bash
> mkdir -p output
> ```

---

## Command Reference

| Command | Description |
|---|---|
| `grep -i "<name>" etc/resources/aws/service-index.csv` | Search for a service ID |
| `xaligo generate excalidraw --xal <xal> -o <out> --services <csv>` | Convert .xal → .excalidraw with legend |
| `xaligo add service --list <csv> --file <excalidraw>` | Add service icons to an existing file |
| `xaligo render <xal> -o <excalidraw>` | Convert .xal → .excalidraw without legend |
