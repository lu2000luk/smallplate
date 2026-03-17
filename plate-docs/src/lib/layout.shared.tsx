import type { BaseLayoutProps } from "fumadocs-ui/layouts/shared";

// fill this with your actual GitHub info, for example:
export const gitConfig = {
  user: "lu2000luk",
  repo: "smallplate",
  branch: "main",
};

export function baseOptions(): BaseLayoutProps {
  return {
    nav: {
      title: "SmallPlate",
    },
    githubUrl: `https://git.lu2000luk.com/${gitConfig.user}/${gitConfig.repo}`,
  };
}
