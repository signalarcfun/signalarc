# SignalArc Enterprise Modular Upgrade Plan

## Core Direction

SignalArc is no longer treated as an MVP.

Target:
- Enterprise-grade product architecture
- Polymarket-class product quality direction
- Mobile-first trading experience
- Modular frontend architecture
- Modular backend architecture
- Professional market discovery
- Reliable portfolio, activity, analytics, and indexer systems
- Production-ready infra, security, and ops discipline

## Contract Policy

Existing smart contracts must remain unchanged in this upgrade track.

Rules:
- Do not modify existing smart contracts.
- Do not introduce smart contract V2 in this roadmap.
- Do not change deployed contract behavior.
- Existing Web3 flows must keep working.
- Contract integration code may be modularized.
- Contracts themselves must not be changed.

## Modular Architecture Rules

Global rules:
1. No monolithic feature files.
2. No mixing UI, API calls, and business logic in one large component.
3. Every product domain must have its own module boundary.
4. Shared constants must live in one canonical place.
5. Frontend and backend must be organized by domain.
6. Large files should be split into components, hooks, services, and utilities.
7. Existing Web3 flows must keep working.
8. Existing contracts must remain unchanged.

## Target Frontend Modules

- markets
- trading
- portfolio
- categories
- wallet
- activity
- analytics
- uploads

## Target Backend Modules

- markets
- portfolio
- activity
- indexer
- analytics
- uploads
- shared/db
- shared/config
- shared/errors
- shared/validation
- shared/logger

## Backend Boundary Rules

- Controller = HTTP boundary
- Service = business logic
- Repository = database access
- Schema = validation
- Types = shared domain contracts

## Phase 0 — Enterprise Audit & Blueprint

Goal:
Create the enterprise blueprint before major refactors.

Scope:
- Audit current frontend structure
- Audit current backend structure
- Audit database and API structure
- Audit indexer and analytics flow
- Audit deployment workflow
- Audit security and ops posture
- Audit UX and product gaps
- Define target modular architecture
- Define implementation sequence

Progress:
- [x] Phase 0 branch created
- [x] Phase 0 blueprint file drafted
- [ ] Current architecture audit completed
- [ ] Target architecture finalized
- [ ] PR opened
- [ ] PR merged

## Phase 1 — Modular Frontend Foundation

Goal:
Refactor frontend into domain modules without changing product behavior.

Scope:
- Create modules/categories
- Create modules/markets
- Create modules/trading
- Create modules/portfolio
- Create modules/activity
- Create modules/wallet
- Move category constants into canonical module
- Split large market components
- Move business logic into hooks/lib files
- Preserve existing routes
- Preserve existing Web3 behavior
- Do not change contracts

Definition of Done:
- Frontend domain modules exist
- Shared constants are centralized
- Large market files are reduced
- Existing market creation and trading flows still work
- pnpm lint:web passes
- pnpm build:web passes

## Phase 2 — Modular Backend Foundation

Goal:
Refactor backend into domain modules with controller/service/repository boundaries.

Scope:
- Move market logic into markets module
- Move portfolio logic into portfolio module
- Move analytics logic into analytics module
- Move upload logic into uploads module
- Move indexer logic into indexer module
- Introduce validation schemas
- Introduce shared error handling
- Introduce shared logger/config patterns
- Preserve existing API behavior

## Phase 3 — Mobile-first Product Design System

Goal:
Make SignalArc feel like a professional mobile-first trading product.

Scope:
- Mobile-first app shell
- Bottom navigation
- Professional dark theme refinement
- Consistent cards, buttons, tabs, and chips
- Better loading, empty, and error states
- Header/navigation cleanup

Primary Mobile Navigation:
- Home
- Markets
- Portfolio
- Activity
- Profile

## Phase 4 — Market Discovery & Category System

Goal:
Build a real market discovery layer.

Official Categories:
- All
- Crypto
- Sports
- Politics
- Macro
- AI
- Tech
- Arc
- Other

Discovery Tabs:
- Live
- Trending
- New
- Ending Soon
- Resolved

Initial Route Model:
- /markets?tab=live&category=crypto
- /markets?tab=trending&category=sports
- /markets?tab=new&category=ai

Scope:
- Standardize market categories
- Update create-market form to use official categories
- Redesign /markets
- Add discovery tabs
- Add category chips
- Add search
- Add filter/sort behavior
- Redesign market cards

## Phase 5 — Market Detail, Trading, Portfolio & Activity UX

Goal:
Turn SignalArc into a usable trading product.

Scope:
- Better market detail header
- Cover image hero
- Clear title/status/category
- YES/NO probability section
- Sticky mobile trade action
- Transaction preview
- Transaction status feedback
- Portfolio summary
- Open positions
- Claimable payouts
- Market activity feed
- Wallet activity feed
- Do not change contracts

## Phase 6 — Backend API & Database Upgrade

Goal:
Support discovery, portfolio, activity, and analytics through backend APIs.

Target APIs:
- GET /markets?category=crypto&status=open&sort=trending
- GET /markets/featured
- GET /markets/:id/activity
- GET /wallets/:address/portfolio
- GET /wallets/:address/positions
- GET /wallets/:address/activity
- GET /analytics/snapshot

Scope:
- Add database indexes
- Normalize category values
- Improve market status querying
- Store or derive activity feed
- Store or derive wallet positions if needed
- Maintain compatibility with current data

## Phase 7 — Indexer & Analytics Reliability

Goal:
Make onchain data ingestion reliable enough for a serious product.

Scope:
- Reliable event indexing
- Last indexed block tracking
- Retry logic
- Idempotent event handling
- Event deduplication
- Backfill command
- Basic reorg awareness
- Transaction failure visibility
- Analytics snapshot automation
- Monitoring/logging for indexer jobs

Contract Rule:
Index existing contracts only. Do not change contract event definitions.

## Phase 8 — Infra, Security, Ops & Production Readiness

Goal:
Make SignalArc operationally credible.

Scope:
- Staging environment
- Production environment hygiene
- Environment separation
- Deployment checklist
- Rollback process
- Secret hygiene
- API validation
- Rate limiting where needed
- Safe upload handling
- Wallet transaction safety messaging
- API health monitoring
- Backend logs
- Indexer logs
- Error tracking
- Uptime monitoring
- Incident response notes
- Release checklist
- QA checklist
- Mobile QA checklist
- User docs
- Developer docs
- API docs
- Operational docs

## Recommended Execution Order

1. Phase 0 — Enterprise Audit & Blueprint
2. Phase 1 — Modular Frontend Foundation
3. Phase 2 — Modular Backend Foundation
4. Phase 4 — Market Discovery & Category System
5. Phase 3 — Mobile-first Product Design System
6. Phase 5 — Trading / Portfolio / Activity UX
7. Phase 6 — Backend API & Database Upgrade
8. Phase 7 — Indexer & Analytics Reliability
9. Phase 8 — Infra / Security / Ops / Production Readiness

## Current Progress Summary

- [x] Phase 0 branch created
- [x] Phase 0 blueprint file drafted
- [ ] Phase 0 audit completed
- [ ] Phase 0 PR opened
- [ ] Phase 0 PR merged
- [ ] Phase 1 — Modular Frontend Foundation
- [ ] Phase 2 — Modular Backend Foundation
- [ ] Phase 3 — Mobile-first Product Design System
- [ ] Phase 4 — Market Discovery & Category System
- [ ] Phase 5 — Market Detail / Trading / Portfolio / Activity UX
- [ ] Phase 6 — Backend API & Database Upgrade
- [ ] Phase 7 — Indexer & Analytics Reliability
- [ ] Phase 8 — Infra / Security / Ops / Production Readiness
