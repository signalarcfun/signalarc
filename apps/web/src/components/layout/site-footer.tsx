import Link from "next/link"

const productLinks = [
  { href: "/markets", label: "Markets", external: false },
  { href: "/portfolio", label: "Portfolio", external: false },
  { href: "/intelligence", label: "Intelligence", external: false },
]

const resourceLinks = [
  {
    href: "https://docs.signalarc.fun",
    label: "Docs",
    external: true,
  },
  {
    href: "https://docs.signalarc.fun/API",
    label: "API Reference",
    external: true,
  },
  {
    href: "https://docs.signalarc.fun/AGENT_API",
    label: "Custom GPT / Agent API",
    external: true,
  },
  {
    href: "https://docs.signalarc.fun/privacy",
    label: "Privacy",
    external: true,
  },
]

const reviewerLinks = [
  {
    href: "https://signalarc.fun",
    label: "Live App",
    external: true,
  },
  {
    href: "https://api.signalarc.fun/health",
    label: "API Health",
    external: true,
  },
  {
    href: "https://docs.signalarc.fun",
    label: "Docs",
    external: true,
  },
  {
    href: "https://docs.signalarc.fun/AGENT_API",
    label: "Agent API Docs",
    external: true,
  },
  {
    href: "https://github.com/wahyu241205/SignalArc",
    label: "GitHub",
    external: true,
  },
]

const linkClassName =
  "text-muted-foreground transition-colors hover:text-foreground"

function FooterLink({
  href,
  label,
  external,
}: {
  href: string
  label: string
  external: boolean
}) {
  if (external) {
    return (
      <a
        href={href}
        target="_blank"
        rel="noopener noreferrer"
        className={linkClassName}
      >
        {label}
      </a>
    )
  }

  return (
    <Link href={href} className={linkClassName}>
      {label}
    </Link>
  )
}

export function SiteFooter() {
  const currentYear = new Date().getFullYear()

  return (
    <footer className="mt-16 border-t border-border/40 bg-background/60">
      <div className="mx-auto w-full max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
        <div className="grid gap-8 sm:grid-cols-2 lg:grid-cols-4">
          <div>
            <div className="flex items-center gap-2">
              <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-indigo-500/20">
                <svg
                  className="h-3.5 w-3.5 text-indigo-400"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  strokeWidth={2}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
                  />
                </svg>
              </div>
              <span className="text-sm font-semibold text-foreground">
                SignalArc
              </span>
            </div>
            <p className="mt-3 text-xs leading-relaxed text-muted-foreground">
              Arc-native prediction market infrastructure for USDC-settled event
              markets.
            </p>
            <p className="mt-3 text-xs font-medium text-indigo-300/80">
              Arc Testnet preview. Testnet assets only. No real funds or
              production settlement.
            </p>
          </div>

          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-foreground">
              Product
            </h3>
            <ul className="mt-3 space-y-2 text-sm">
              {productLinks.map((link) => (
                <li key={`product-${link.href}`}>
                  <FooterLink
                    href={link.href}
                    label={link.label}
                    external={link.external}
                  />
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-foreground">
              Resources
            </h3>
            <ul className="mt-3 space-y-2 text-sm">
              {resourceLinks.map((link) => (
                <li key={`resource-${link.href}-${link.label}`}>
                  <FooterLink
                    href={link.href}
                    label={link.label}
                    external={link.external}
                  />
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-foreground">
              For Reviewers
            </h3>
            <ul className="mt-3 space-y-2 text-sm">
              {reviewerLinks.map((link) => (
                <li key={`reviewer-${link.href}-${link.label}`}>
                  <FooterLink
                    href={link.href}
                    label={link.label}
                    external={link.external}
                  />
                </li>
              ))}
            </ul>
          </div>
        </div>

        <div className="mt-10 flex flex-col gap-2 border-t border-border/40 pt-6 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
          <p>© {currentYear} SignalArc. Arc Testnet prototype.</p>
          <p>
            Built on{" "}
            <a
              href="https://docs.arc.io/"
              target="_blank"
              rel="noopener noreferrer"
              className="text-indigo-400 transition-colors hover:text-indigo-300"
            >
              Arc
            </a>{" "}
            with{" "}
            <a
              href="https://developers.circle.com/"
              target="_blank"
              rel="noopener noreferrer"
              className="text-indigo-400 transition-colors hover:text-indigo-300"
            >
              Circle
            </a>
            .
          </p>
        </div>
      </div>
    </footer>
  )
}
