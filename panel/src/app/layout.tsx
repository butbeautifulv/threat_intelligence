import "./globals.css";

export const metadata = {
  title: "Threat Intelligence Graph",
  description: "Obsidian-style graph + markdown explorer over Neo4j",
} as const;

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
