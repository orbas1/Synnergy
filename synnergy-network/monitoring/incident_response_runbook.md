# Synnergy Incident Response Runbook

## Service Level Objectives

- **Node API availability:** 99.9% monthly uptime
- **CLI command latency:** 95th percentile below 500ms
- **GUI load time:** 95th percentile below 2s

## Roles and Responsibilities

- **On-call engineer:** First responder responsible for triage and mitigation.
- **Incident commander:** Coordinates response once severity is confirmed.
- **Communications lead:** Handles stakeholder updates and public notices.
- **SRE manager:** Oversees incident reviews and ensures follow-up completion.

## Detection and Triage

1. Alert triggers via Prometheus Alertmanager or external monitoring service.
2. On-call reviews Grafana dashboards and correlates logs in the HealthLogger output.
3. Classify severity (SEV1–SEV3) based on user impact and SLO breach.
   - **SEV1:** Total outage or data loss.
   - **SEV2:** Degraded performance or partial outage.
   - **SEV3:** Minor issue with limited user impact.
4. If SEV1 or SEV2, page the incident commander immediately.

## Escalation Process

1. Incident commander assembles response team on dedicated channel.
2. Mitigate immediate user impact: restart services, scale nodes, or roll back deployments.
3. Collect artefacts—logs, metrics snapshots and timelines—for later analysis.
4. Communications lead publishes status updates every 30 minutes until resolution.
5. SRE manager schedules postmortem review within 24 hours.

## Post-Incident

- Publish a blameless postmortem within five business days.
- Update monitoring dashboards and SLOs as needed.
- Track follow-up actions to completion and verify in the next sprint review.
