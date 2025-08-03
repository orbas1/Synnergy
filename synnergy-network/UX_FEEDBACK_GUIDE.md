# User Feedback and UX Guide

This guide outlines how to collect feedback from the community and translate it into user‑experience improvements for both the GUI projects and the CLI tools.

## Feedback Channels
- **GitHub Issues** – use the [UX Feedback template](../.github/ISSUE_TEMPLATE/ux_feedback.md) for bug reports or enhancement requests.
- **Discussion Forum** – raise broader UX ideas in the community forum or Discord.
- **User Interviews** – schedule periodic calls with power users to capture qualitative insights.

## Gathering Feedback
- Encourage users to describe their workflow, pain points and desired enhancements.
- Tag issues with `ux` and the relevant component label (e.g. `wallet`, `cli`).
- Collect environment details and screenshots to help reproduce problems.

## Triage and Prioritisation
- Review new UX issues in the weekly triage meeting.
- Classify each issue by impact (`low`, `medium`, `high`) and assign an owner.
- Link related issues or design documents to provide context.

## Feedback Lifecycle
1. **Collect** – capture reports from GitHub, forums and interviews.
2. **Triage** – prioritise and assign owners in the weekly meeting.
3. **Implement** – develop and review fixes or enhancements.
4. **Validate** – verify changes with the original reporter or usability tests.
5. **Close** – merge the change, update release notes and mark the issue resolved.

## Usability Testing
- Conduct task‑based testing sessions for new features in GUI modules and CLI commands.
- Observe users completing common tasks and note friction or confusion.
- Capture timing metrics, error rates and qualitative comments for each session.

## Accessibility Review
- Verify GUI components against WCAG 2.1 AA guidelines including:
  - 4.5:1 colour contrast for text
  - Navigable focus order and keyboard shortcuts
  - ARIA labels for interactive elements
  - Alt text for images and icons
- For CLI tools, ensure commands provide clear help output and avoid colour‑only cues.
- Test interfaces with screen readers, high‑contrast themes and limited bandwidth scenarios.

## UX Metrics and Reporting
- Track task completion rates, error counts and satisfaction scores after each release.
- Monitor accessibility compliance using automated tools and manual audits.
- Summarise key metrics and closed issues in the monthly community call.

## Onboarding and Interface Refinement
- Document first‑run experiences for wallets, explorers and CLI setup commands.
- Provide contextual help and prompts that guide users through initial configuration.
- Iterate on layout and terminology based on user feedback to reduce learning curves.

## Tracking Improvements
- Link UX issues to pull requests that address them.
- Maintain a changelog section highlighting usability and accessibility fixes.
- Periodically review open feedback to prioritise future enhancements.

