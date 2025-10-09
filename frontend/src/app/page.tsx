import Link from 'next/link';
import { ArrowRight, Users, Calendar, Award, BarChart3 } from 'lucide-react';

export default function Home() {
  return (
    <div className="flex flex-col min-h-screen">
      {/* Header */}
      <header className="border-b">
        <div className="container mx-auto px-4 py-4 flex justify-between items-center">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center">
              <span className="text-white font-bold text-xl">V</span>
            </div>
            <span className="text-xl font-bold">VolunteerSync</span>
          </div>
          <nav className="hidden md:flex gap-6">
            <Link href="/opportunities" className="text-sm font-medium hover:text-primary">
              Find Opportunities
            </Link>
            <Link href="/organizations" className="text-sm font-medium hover:text-primary">
              Organizations
            </Link>
            <Link href="/login" className="text-sm font-medium hover:text-primary">
              Log In
            </Link>
            <Link
              href="/register"
              className="px-4 py-2 bg-primary text-white text-sm font-medium rounded-md hover:bg-primary/90"
            >
              Get Started
            </Link>
          </nav>
        </div>
      </header>

      {/* Hero Section */}
      <section className="flex-1 flex items-center justify-center py-20 px-4">
        <div className="container mx-auto text-center max-w-4xl">
          <h1 className="text-4xl md:text-6xl font-bold tracking-tight mb-6">
            Connect, Volunteer,
            <span className="text-primary"> Make an Impact</span>
          </h1>
          <p className="text-lg md:text-xl text-muted-foreground mb-8 max-w-2xl mx-auto">
            VolunteerSync streamlines volunteer management for organizations and makes it easy for
            volunteers to discover meaningful opportunities in their community.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/register?type=volunteer"
              className="inline-flex items-center justify-center px-6 py-3 bg-primary text-white font-medium rounded-lg hover:bg-primary/90 transition-colors"
            >
              I'm a Volunteer
              <ArrowRight className="ml-2 h-5 w-5" />
            </Link>
            <Link
              href="/register?type=organization"
              className="inline-flex items-center justify-center px-6 py-3 border-2 border-primary text-primary font-medium rounded-lg hover:bg-primary/5 transition-colors"
            >
              I'm an Organization
              <ArrowRight className="ml-2 h-5 w-5" />
            </Link>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 bg-muted/30">
        <div className="container mx-auto px-4">
          <h2 className="text-3xl font-bold text-center mb-12">Why VolunteerSync?</h2>
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
            <div className="flex flex-col items-center text-center">
              <div className="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center mb-4">
                <Users className="h-6 w-6 text-primary" />
              </div>
              <h3 className="font-semibold mb-2">Easy Registration</h3>
              <p className="text-sm text-muted-foreground">
                Find and register for volunteer opportunities in under 3 minutes
              </p>
            </div>
            <div className="flex flex-col items-center text-center">
              <div className="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center mb-4">
                <Calendar className="h-6 w-6 text-primary" />
              </div>
              <h3 className="font-semibold mb-2">Smart Scheduling</h3>
              <p className="text-sm text-muted-foreground">
                Organizations can post events and manage volunteers effortlessly
              </p>
            </div>
            <div className="flex flex-col items-center text-center">
              <div className="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center mb-4">
                <Award className="h-6 w-6 text-primary" />
              </div>
              <h3 className="font-semibold mb-2">Track Your Impact</h3>
              <p className="text-sm text-muted-foreground">
                Log hours and earn achievements as you make a difference
              </p>
            </div>
            <div className="flex flex-col items-center text-center">
              <div className="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center mb-4">
                <BarChart3 className="h-6 w-6 text-primary" />
              </div>
              <h3 className="font-semibold mb-2">Powerful Analytics</h3>
              <p className="text-sm text-muted-foreground">
                Organizations get insights and reporting to measure success
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t py-8">
        <div className="container mx-auto px-4 text-center text-sm text-muted-foreground">
          <p>&copy; 2025 VolunteerSync. All rights reserved.</p>
        </div>
      </footer>
    </div>
  );
}
