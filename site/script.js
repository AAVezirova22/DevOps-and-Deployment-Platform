const tabs = document.querySelectorAll(".tab");
const panels = document.querySelectorAll(".tab-panel");
const root = document.documentElement;

root.classList.add("js");

const header = document.querySelector("[data-site-header]");
const navLinks = document.querySelectorAll("[data-nav-link]");
const navIndicator = document.querySelector(".nav-indicator");
const mobileNav = document.querySelector("[data-mobile-nav]");
const navToggle = document.querySelector(".nav-toggle");
const revealItems = document.querySelectorAll(".reveal, .reveal-group");

const setScrollState = () => {
  const maxScroll = document.body.scrollHeight - window.innerHeight;
  const progress = maxScroll > 0 ? window.scrollY / maxScroll : 0;

  root.style.setProperty("--scroll-progress", progress.toFixed(4));
  header?.classList.toggle("is-scrolled", window.scrollY > 16);
};

const moveNavIndicator = (activeLink) => {
  if (!activeLink || !navIndicator) {
    return;
  }

  const navRect = activeLink.parentElement.getBoundingClientRect();
  const linkRect = activeLink.getBoundingClientRect();

  root.style.setProperty("--nav-indicator-x", `${linkRect.left - navRect.left - 5}px`);
  root.style.setProperty("--nav-indicator-width", `${linkRect.width}px`);
  root.style.setProperty("--nav-indicator-opacity", "1");
};

const setActiveNav = (sectionId) => {
  navLinks.forEach((link) => {
    const isActive = link.getAttribute("href") === `#${sectionId}`;

    link.classList.toggle("is-active", isActive);
    if (isActive) {
      moveNavIndicator(link);
    }
  });
};

revealItems.forEach((item, index) => {
  item.style.setProperty("--reveal-delay", `${Math.min(index * 55, 220)}ms`);

  if (item.classList.contains("reveal-group")) {
    Array.from(item.children).forEach((child, childIndex) => {
      child.style.setProperty("--reveal-delay", `${Math.min(childIndex * 75, 300)}ms`);
    });
  }
});

tabs.forEach((tab) => {
  tab.addEventListener("click", () => {
    const target = tab.dataset.tab;

    tabs.forEach((item) => item.classList.toggle("is-active", item === tab));
    panels.forEach((panel) => panel.classList.toggle("is-active", panel.id === `tab-${target}`));
  });
});

navLinks.forEach((link) => {
  link.addEventListener("mouseenter", () => moveNavIndicator(link));
  link.addEventListener("mouseleave", () => {
    const activeLink = document.querySelector("[data-nav-link].is-active");
    moveNavIndicator(activeLink);
  });
});

navToggle?.addEventListener("click", () => {
  const isOpen = navToggle.getAttribute("aria-expanded") === "true";

  navToggle.setAttribute("aria-expanded", String(!isOpen));
  navToggle.setAttribute("aria-label", isOpen ? "Open navigation" : "Close navigation");
  mobileNav?.classList.toggle("is-open", !isOpen);
  root.classList.toggle("nav-open", !isOpen);
});

mobileNav?.querySelectorAll("a").forEach((link) => {
  link.addEventListener("click", () => {
    navToggle?.setAttribute("aria-expanded", "false");
    navToggle?.setAttribute("aria-label", "Open navigation");
    mobileNav.classList.remove("is-open");
    root.classList.remove("nav-open");
  });
});

document.querySelectorAll(".copy-button").forEach((button) => {
  button.addEventListener("click", async () => {
    const text = button.dataset.copy;

    try {
      await navigator.clipboard.writeText(text);
      button.textContent = "Copied";
      button.classList.add("is-copied");
      setTimeout(() => {
        button.textContent = "Copy";
        button.classList.remove("is-copied");
      }, 1400);
    } catch {
      button.textContent = "Select";
    }
  });
});

if ("IntersectionObserver" in window) {
  const revealObserver = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add("is-visible");
          revealObserver.unobserve(entry.target);
        }
      });
    },
    { rootMargin: "0px 0px -12% 0px", threshold: 0.14 },
  );

  revealItems.forEach((item) => revealObserver.observe(item));

  const sectionObserver = new IntersectionObserver(
    (entries) => {
      const visibleEntry = entries
        .filter((entry) => entry.isIntersecting)
        .sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0];

      if (visibleEntry?.target.id) {
        setActiveNav(visibleEntry.target.id);
      }
    },
    { rootMargin: "-28% 0px -48% 0px", threshold: [0.2, 0.42, 0.64] },
  );

  document.querySelectorAll("[data-section][id]").forEach((section) => sectionObserver.observe(section));
} else {
  revealItems.forEach((item) => item.classList.add("is-visible"));
}

window.addEventListener("scroll", setScrollState, { passive: true });
window.addEventListener("resize", () => {
  setScrollState();
  moveNavIndicator(document.querySelector("[data-nav-link].is-active"));
});

setScrollState();
setActiveNav(window.location.hash.replace("#", "") || "install");
