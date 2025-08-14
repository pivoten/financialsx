import * as React from "react"
import { cn } from "../../lib/utils"

const Sidebar = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn(
        "flex h-full w-64 flex-col border-r bg-card text-card-foreground shadow-sm",
        className
      )}
      {...props}
    />
  )
)
Sidebar.displayName = "Sidebar"

const SidebarHeader = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn("flex h-16 items-center border-b px-6", className)}
      {...props}
    />
  )
)
SidebarHeader.displayName = "SidebarHeader"

const SidebarContent = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn("flex-1 overflow-auto py-4", className)}
      {...props}
    />
  )
)
SidebarContent.displayName = "SidebarContent"

const SidebarNav = React.forwardRef<HTMLElement, React.HTMLAttributes<HTMLElement>>(
  ({ className, ...props }, ref) => (
    <nav
      ref={ref}
      className={cn("space-y-1 px-3", className)}
      {...props}
    />
  )
)
SidebarNav.displayName = "SidebarNav"

export interface SidebarNavItemProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  active?: boolean
}

const SidebarNavItem = React.forwardRef<HTMLAnchorElement, SidebarNavItemProps>(
  ({ className, active, ...props }, ref) => (
    <a
      ref={ref}
      className={cn(
        "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground",
        active && "bg-accent text-accent-foreground",
        className
      )}
      {...props}
    />
  )
)
SidebarNavItem.displayName = "SidebarNavItem"

export interface SidebarNavGroupProps extends React.HTMLAttributes<HTMLDivElement> {
  title?: string
  children?: React.ReactNode
}

const SidebarNavGroup = React.forwardRef<HTMLDivElement, SidebarNavGroupProps>(
  ({ className, title, children, ...props }, ref) => (
    <div ref={ref} className={cn("px-3", className)} {...props}>
      {title && (
        <h4 className="mb-2 px-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          {title}
        </h4>
      )}
      <div className="space-y-1">{children}</div>
    </div>
  )
)
SidebarNavGroup.displayName = "SidebarNavGroup"

export { Sidebar, SidebarHeader, SidebarContent, SidebarNav, SidebarNavItem, SidebarNavGroup }
