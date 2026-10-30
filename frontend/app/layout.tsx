import "./globals.css";
export const metadata = {
  title: "CourtFlow | Find your next game",
  description: "Fictional badminton booking demo. No real payments.",
};
export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
