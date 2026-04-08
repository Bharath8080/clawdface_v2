import { Nav } from "@/components/landing/Nav";
import { Footer } from "@/components/landing/Footer";

export default function PrivacyPolicyPage() {
  return (
    <div className="min-h-screen bg-[#050505] text-[#E0E0E0] lg:font-inter selection:bg-[#82e8b2] selection:text-black overflow-x-hidden flex flex-col">
      <Nav />
      <main className="flex-1 pt-32 pb-20 px-6 max-w-4xl mx-auto w-full relative z-10">
        <h1 className="text-4xl md:text-5xl font-bold tracking-tight text-white mb-10">Privacy Policy</h1>
        
        <div className="text-[15px] text-zinc-300 leading-relaxed">
          <h2 className="text-2xl font-bold text-white mt-12 mb-4">Introduction</h2>
          <p className="mb-4">
            At ClawdFace ("we," "us," or "our"), we take your privacy seriously. This Privacy Policy explains how we collect, use, share, and protect your personal information when you interact with our AI-powered video generation platform and related services (collectively, the "Platform").
          </p>
          <p className="mb-4">
            Our Platform enables businesses to create personalized AI-generated videos at scale. This policy applies to information we collect directly from you and automatically through your use of our Platform.
          </p>
          <p className="mb-8">
            <strong className="text-white">Important Note:</strong> This policy does not cover information we process on behalf of our business clients. If you have questions about data processing by one of our business clients, please contact them directly.
          </p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">What Information We Collect</h2>
          
          <h3 className="text-[15px] font-semibold text-white mb-2">Information You Provide Directly</h3>
          <ul className="list-disc pl-5 mb-6 space-y-2 marker:text-zinc-500">
            <li><strong className="text-white">Account Information:</strong> Name, email address, phone number, company details, professional title, contact preferences, credentials, and profile information</li>
            <li><strong className="text-white">Content and Media:</strong> Audio/video samples (including biometric identifiers), scripts, logos, branding, generated video content, metadata</li>
            <li><strong className="text-white">Communication Data:</strong> Messages, support requests, survey responses, testimonials</li>
            <li><strong className="text-white">Payment Information:</strong> Billing details, payment card info (via third-party processors), transaction history</li>
            <li><strong className="text-white">Job Application Data:</strong> Resume, cover letter, credentials, work history, education, diversity info</li>
          </ul>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">Information Collected Automatically</h3>
          <ul className="list-disc pl-5 mb-6 space-y-2 marker:text-zinc-500">
            <li><strong className="text-white">Technical Data:</strong> Device info, usage patterns, error logs, IP, location</li>
            <li><strong className="text-white">Analytics Information:</strong> Page views, sessions, navigation paths, A/B tests</li>
          </ul>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">Information from Third Parties</h3>
          <ul className="list-disc pl-5 mb-8 space-y-2 marker:text-zinc-500">
            <li><strong className="text-white">Integration Partners:</strong> Social media, CRM, SSO providers</li>
            <li><strong className="text-white">Public Sources:</strong> Business info, networking profiles, directories</li>
          </ul>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">User's right to their data</h2>
          
          <h3 className="text-[15px] font-semibold text-white mb-2">Right to Data Deletion (Right to be Forgotten)</h3>
          <ul className="list-disc pl-5 mb-6 space-y-2 marker:text-zinc-500">
            <li>Users have the right to request deletion of their personal data at any time in accordance with Article 17 of the GDPR.</li>
          </ul>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">How Users Can Request Deletion - Submitting a Data Deletion Request</h3>
          <ul className="list-disc pl-5 mb-6 space-y-2 marker:text-zinc-500">
            <li>Contacting us at support@clawdface.io with the subject line “Data Deletion Request”.</li>
          </ul>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">Identity Verification - Verification of Requests</h3>
          <ul className="list-disc pl-5 mb-6 space-y-2 marker:text-zinc-500">
            <li>To protect user privacy and prevent unauthorized deletion, we may verify the identity of the requester before processing any deletion request.</li>
          </ul>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">Scope of Data Deleted - Data Covered by Deletion</h3>
          <p className="mb-2">Upon successful verification, we will permanently delete or anonymize personal data including, but not limited to:</p>
          <ul className="list-disc pl-5 mb-6 space-y-2 marker:text-zinc-500">
            <li>Account information</li>
            <li>User-generated content</li>
            <li>Usage and activity data</li>
            <li>Stored preferences</li>
          </ul>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">Timeline for Deletion - Processing Time</h3>
          <ul className="list-disc pl-5 mb-6 space-y-2 marker:text-zinc-500">
            <li>We will process valid deletion requests within 30 days, as required under GDPR. If additional time is required due to technical or legal reasons, users will be informed accordingly.</li>
          </ul>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">Legal & Regulatory Exceptions - Exceptions to Deletion</h3>
          <p className="mb-2">Certain data may be retained where required to:</p>
          <ul className="list-disc pl-5 mb-4 space-y-2 marker:text-zinc-500">
            <li>Comply with legal or regulatory obligations</li>
            <li>Resolve disputes or enforce agreements</li>
            <li>Maintain financial, tax, or audit records</li>
            <li>Prevent fraud, abuse, or security incidents</li>
          </ul>
          <p className="mb-6">Such data will be securely restricted and not used for any other purpose.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">Backups & Logs - Backups and Residual Data</h3>
          <ul className="list-disc pl-5 mb-6 space-y-2 marker:text-zinc-500">
            <li>Data may persist in encrypted backups for a limited period. These backups are securely isolated and automatically deleted or overwritten in accordance with our retention policy.</li>
          </ul>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">Confirmation to User - Confirmation of Deletion</h3>
          <ul className="list-disc pl-5 mb-8 space-y-2 marker:text-zinc-500">
            <li>Once the deletion process is complete, users will receive confirmation that their personal data has been deleted or anonymized.</li>
          </ul>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">How We Use Your Information</h2>
          <ul className="list-disc pl-5 mb-8 space-y-3 marker:text-zinc-500">
            <li><strong className="text-white">Core Platform Services:</strong> Video generation, account management, support, security</li>
            <li><strong className="text-white">Business Operations:</strong> Service improvement, AI model training (anonymized), monitoring, compliance</li>
            <li><strong className="text-white">Marketing & Communications:</strong> Updates, promotions (with consent), personalization, analytics</li>
            <li><strong className="text-white">Research & Development:</strong> Innovation, industry analysis, QA, academic research (anonymized)</li>
          </ul>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">How We Share Your Information</h2>
          <ul className="list-disc pl-5 mb-8 space-y-3 marker:text-zinc-500">
            <li><strong className="text-white">Service Providers:</strong> Hosting, payments, email, analytics, support tools</li>
            <li><strong className="text-white">Business Partners:</strong> Integration, resellers, marketing partners</li>
            <li><strong className="text-white">Legal Requirements:</strong> Disclosure to comply with law, protect rights, investigate fraud</li>
            <li><strong className="text-white">Business Transactions:</strong> Data transfer during mergers or acquisitions</li>
          </ul>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">Biometric Data Notice</h2>
          <h3 className="text-[15px] font-semibold text-white mb-2">What We Collect</h3>
          <ul className="list-disc pl-5 mb-6 space-y-2 marker:text-zinc-500">
            <li>Voice patterns from audio</li>
            <li>Facial geometry from video</li>
            <li>Expression data for rendering</li>
          </ul>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">How We Use Biometric Data</h3>
          <ul className="list-disc pl-5 mb-6 space-y-2 marker:text-zinc-500">
            <li>Generate personalized AI videos</li>
            <li>Train AI models (anonymized)</li>
            <li>Ensure quality and authenticity</li>
            <li>Comply with fraud/security requirements</li>
          </ul>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">Protection and Retention</h3>
          <ul className="list-disc pl-5 mb-8 space-y-2 marker:text-zinc-500">
            <li>Industry-standard security</li>
            <li>Retention only as long as needed</li>
            <li>User-requested deletion anytime</li>
            <li>No selling/renting biometric data</li>
          </ul>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">Your Privacy Rights</h2>
          <h3 className="text-[15px] font-semibold text-white mb-2">Access and Control</h3>
          <ul className="list-disc pl-5 mb-6 space-y-2 marker:text-zinc-500">
            <li>Request copy of personal data</li>
            <li>Update/correct account details</li>
            <li>Export data</li>
            <li>Request account deletion</li>
          </ul>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">Communication Preferences</h3>
          <ul className="list-disc pl-5 mb-6 space-y-2 marker:text-zinc-500">
            <li>Opt-out of marketing</li>
            <li>Manage notifications</li>
            <li>Control cookies</li>
          </ul>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">Data Protection Rights</h3>
          <ul className="list-disc pl-5 mb-8 space-y-2 marker:text-zinc-500">
            <li>Right to rectification, erasure, restrict processing</li>
            <li>Right to data portability</li>
            <li>Right to object</li>
          </ul>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">Data Security and Storage</h2>
          <ul className="list-disc pl-5 mb-8 space-y-3 marker:text-zinc-500">
            <li><strong className="text-white">Security Measures:</strong> Encryption, access controls, audits, training, incident response</li>
            <li><strong className="text-white">Data Retention:</strong> Account data (active + reasonable time), biometric (1 year or request), analytics (anonymized), legal obligations</li>
          </ul>

          <h2 className="text-2xl font-bold text-white mt-12 mb-4">International Transfers</h2>
          <p className="mb-8">
            Our Platform operates globally, and your data may be transferred to and processed in countries other than your own. We ensure appropriate safeguards are in place for all international transfers.
          </p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">Cookies and Tracking Technologies</h2>
          <ul className="list-disc pl-5 mb-4 space-y-3 marker:text-zinc-500">
            <li><strong className="text-white">Essential Cookies:</strong> Authentication, session, fraud prevention</li>
            <li><strong className="text-white">Analytics Cookies:</strong> Usage tracking, feature adoption, error monitoring</li>
            <li><strong className="text-white">Marketing Cookies:</strong> Ads, social media, campaign attribution</li>
          </ul>
          <p className="mb-8">You can control cookies via browser settings, preference center, or third-party tools.</p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-4">Children's Privacy</h2>
          <p className="mb-8">
            Our Platform is designed for business use and not for individuals under 18. We do not knowingly collect personal information from children. If discovered, such information will be deleted promptly.
          </p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-4">Changes to This Policy</h2>
          <p className="mb-8">
            We may update this Privacy Policy periodically. Material changes will be posted, emailed to registered users, and provided with reasonable review time before taking effect.
          </p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">International Users</h2>
          <h3 className="text-[15px] font-semibold text-white mb-2">European Users (GDPR)</h3>
          <p className="mb-6">Additional rights and protections apply under GDPR. Contact our Data Protection Officer at <a className="text-[#82e8b2] hover:text-[#6bd69e] hover:underline transition-colors" href="mailto:privacy@clawdface.io">privacy@clawdface.io</a> to exercise rights.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">California Users (CCPA)</h3>
          <p className="mb-8">California residents have rights under the CCPA. Please refer to our California Privacy Notice for details.</p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">Contact Us</h2>
          <ul className="list-disc pl-5 mb-8 space-y-3 marker:text-zinc-500">
            <li><strong className="text-white">Privacy Questions:</strong> <a className="text-[#82e8b2] hover:text-[#6bd69e] hover:underline transition-colors" href="mailto:privacy@clawdface.io">privacy@clawdface.io</a></li>
            <li><strong className="text-white">General Support:</strong> <a className="text-[#82e8b2] hover:text-[#6bd69e] hover:underline transition-colors" href="mailto:support@clawdface.io">support@clawdface.io</a></li>
            <li><strong className="text-white">Data Protection Officer (EU/UK):</strong> <a className="text-[#82e8b2] hover:text-[#6bd69e] hover:underline transition-colors" href="mailto:privacy@clawdface.io">privacy@clawdface.io</a></li>
          </ul>
        </div>
      </main>
      <Footer />
    </div>
  );
}
