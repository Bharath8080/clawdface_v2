import { Nav } from "@/components/landing/Nav";
import { Footer } from "@/components/landing/Footer";

export default function TermsOfServicePage() {
  return (
    <div className="min-h-screen bg-canvas text-[#E0E0E0] lg:font-inter selection:bg-[#82e8b2] selection:text-black overflow-x-hidden flex flex-col">
      <Nav />
      <main className="flex-1 pt-32 pb-20 px-6 max-w-4xl mx-auto w-full relative z-10">
        <h1 className="text-4xl md:text-5xl font-bold tracking-tight text-white mb-10">Terms of Service</h1>
        
        <div className="text-[15px] text-zinc-300 leading-relaxed">
          <p className="mb-8">
            These Terms of Service (these "Terms") govern your access to and use of the ClawdFace Platform and related services provided by ClawdFace Inc. ("ClawdFace", "we", "us" or "our"). By accessing or using our services, you agree to be bound by these Terms.
          </p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">1. Definitions</h2>
          <ul className="list-disc pl-5 mb-8 space-y-3 marker:text-zinc-500">
            <li><strong className="text-white">Access Protocols</strong> means the API keys, passwords, access codes, technical specifications, connectivity standards or protocols, or other relevant procedures necessary to access and use the ClawdFace Platform.</li>
            <li><strong className="text-white">AI Services</strong> means third-party machine learning, artificial intelligence systems, platforms, models, and algorithms that may be incorporated into or accessible through the ClawdFace Services.</li>
            <li><strong className="text-white">Authorized User</strong> means each of your employees, agents, independent contractors, and service providers who are authorized to access and use the ClawdFace Platform pursuant to your rights under this Agreement.</li>
            <li><strong className="text-white">Connected Applications</strong> means any system or application owned or controlled by you that is connected to the ClawdFace Services.</li>
            <li><strong className="text-white">Conversational Video Agent</strong> means an AI-powered avatar accessible via the ClawdFace Solution that can engage in human-like conversations with End Users.</li>
            <li><strong className="text-white">Customer Content</strong> means all data, video, audio, text, prompts, images, and other content or materials submitted to the ClawdFace Platform by you or on your behalf, including content from End Users.</li>
            <li><strong className="text-white">Documentation</strong> means the technical materials provided by ClawdFace describing the use and operation of the ClawdFace Platform.</li>
            <li><strong className="text-white">End Users</strong> means individuals who interact with or access your products or services that integrate with the ClawdFace Platform.</li>
            <li><strong className="text-white">ClawdFace Platform</strong> means our software-as-a-service offering that includes interactive media services, made available through our website and related interfaces such as APIs or developer tools.</li>
            <li><strong className="text-white">ClawdFace Services</strong> means the ClawdFace Platform and all related services, including the ClawdFace Solution and Connectors.</li>
            <li><strong className="text-white">ClawdFace Solution</strong> means the Conversational Video Agents and conversation engine for generating human-like conversations.</li>
          </ul>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">2. Accounts and Access</h2>
          <h3 className="text-[15px] font-semibold text-white mb-2">2.1 Account Creation</h3>
          <p className="mb-6">To access certain features, you may need to register an account or connect through a valid third-party service account. You must provide true, accurate, current, and complete information and maintain this information's accuracy.</p>
          
          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">2.2 Account Ownership</h3>
          <p className="mb-6">You are responsible for maintaining the confidentiality of your account credentials and for all activities that occur under your account. You agree to notify ClawdFace immediately of any unauthorized access or security breaches.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">2.3 Access Rights</h3>
          <p className="mb-4">Subject to your compliance with these Terms, ClawdFace grants you a limited, non-exclusive, non-transferable license during the Term to:</p>
          <ul className="list-disc pl-5 mb-8 space-y-2 marker:text-zinc-500">
            <li>Access and use the ClawdFace Platform for personal use, internal business purposes, or incorporation into your own products or services</li>
            <li>Use the Documentation to support your authorized use</li>
            <li>Integrate ClawdFace APIs into your websites and applications for use by End Users</li>
          </ul>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">3. Customer Responsibilities</h2>
          <h3 className="text-[15px] font-semibold text-white mb-2">3.1 Content Responsibility</h3>
          <p className="mb-4">You are solely responsible for all Customer Content, including ensuring you have obtained all necessary rights, licenses, consents, and permissions to submit such content. This includes:</p>
          <ul className="list-disc pl-5 mb-6 space-y-2 marker:text-zinc-500">
            <li>Explicit consents for processing biometric data</li>
            <li>Written releases from individuals whose likeness, voice, or biometric identifiers are included</li>
            <li>Compliance with applicable data privacy laws</li>
          </ul>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">3.2 Generated Content Use</h3>
          <p className="mb-4">You are solely responsible for all uses of content generated through the ClawdFace Platform. You must ensure that generated content does not:</p>
          <ul className="list-disc pl-5 mb-6 space-y-2 marker:text-zinc-500">
            <li>Impersonate individuals without consent</li>
            <li>Misrepresent affiliation, sponsorship, or authorship</li>
            <li>Violate third-party rights</li>
            <li>Cause harm, offense, or confusion</li>
          </ul>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">3.3 AI Transparency</h3>
          <p className="mb-6">You must clearly inform End Users that they are interacting with an AI-generated avatar and comply with all applicable laws regarding synthetic audio or video content.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">3.4 Acceptable Use</h3>
          <p className="mb-4">You agree to comply with our Acceptable Use Policy and ensure that all Customer Content and use of our services:</p>
          <ul className="list-disc pl-5 mb-8 space-y-2 marker:text-zinc-500">
            <li>Is not deceptive, defamatory, obscene, pornographic, or unlawful</li>
            <li>Does not contain malicious code or viruses</li>
            <li>Does not violate third-party rights</li>
          </ul>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">4. Intellectual Property</h2>
          <h3 className="text-[15px] font-semibold text-white mb-2">4.1 ClawdFace Ownership</h3>
          <p className="mb-6">ClawdFace retains all rights, title, and interest in the ClawdFace Platform, Documentation, and related technology. All rights not expressly granted are reserved.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">4.2 Customer Ownership</h3>
          <p className="mb-6">You retain ownership of your Customer Content and any content generated through your use of the ClawdFace Platform, subject to the licenses granted to ClawdFace herein.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">4.3 License to ClawdFace</h3>
          <p className="mb-4">You grant ClawdFace a non-exclusive, worldwide, royalty-free license to use Customer Content to:</p>
          <ul className="list-disc pl-5 mb-6 space-y-2 marker:text-zinc-500">
            <li>Generate requested outputs and provide platform functionality</li>
            <li>Maintain and improve ClawdFace's products and services</li>
            <li>Use aggregated and de-identified usage data for analytics and improvements</li>
          </ul>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">4.4 Feedback</h3>
          <p className="mb-8">Any feedback, suggestions, or recommendations you provide regarding the ClawdFace Platform may be used by ClawdFace without restriction or compensation.</p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">5. Fees and Payments</h2>
          <h3 className="text-[15px] font-semibold text-white mb-2">5.1 Payment Processing</h3>
          <p className="mb-6">We use third-party payment processors including Stripe, Inc. You agree to their terms and authorize us to share payment information as necessary to complete transactions.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">5.2 Payment Terms</h3>
          <p className="mb-6">You shall pay all fees in accordance with the billing terms in effect when fees are due. All fees are non-refundable except as expressly provided in these Terms.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">5.3 Subscriptions and Auto-Renewal</h3>
          <p className="mb-6">If you purchase a subscription, it will automatically renew at our then-current pricing unless you cancel before the renewal date. You may cancel through your account settings.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">5.4 Free Trials</h3>
          <p className="mb-6">Free trials automatically convert to paid subscriptions unless canceled before the trial period ends. By starting a free trial, you agree to this automatic conversion.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">5.5 Taxes</h3>
          <p className="mb-8">Fees are exclusive of all applicable taxes, duties, and charges, which are your responsibility.</p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">6. Restrictions</h2>
          <ul className="list-disc pl-5 mb-8 space-y-3 marker:text-zinc-500">
            <li>Use the services for any unlawful purpose or in violation of these Terms</li>
            <li>Reverse engineer, decompile, or attempt to derive source code</li>
            <li>Access or use the services to build competing products</li>
            <li>Remove or alter proprietary notices</li>
            <li>Share API keys or access credentials with unauthorized parties</li>
            <li>Use the services in ways that could harm our systems or other users</li>
            <li>Enable interactions with Conversational Video Agents without clearly disclosing their AI nature</li>
          </ul>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">7. Third-Party Services</h2>
          <p className="mb-8">Our services may incorporate or provide access to third-party AI services and applications. We are not responsible for these third-party services, and your use is subject to their respective terms and policies.</p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">8. Data and Privacy</h2>
          <h3 className="text-[15px] font-semibold text-white mb-2">8.1 Data Processing</h3>
          <p className="mb-6">Processing of personal data is governed by our Data Processing Addendum and Privacy Policy. We implement appropriate technical and organizational measures to protect your data.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">8.2 Data Retention</h3>
          <p className="mb-6">Inputs and interactions are generally processed on a stateless basis with minimal retention, except where necessary for service functionality or legal compliance.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">8.3 No Training Use</h3>
          <p className="mb-8">We will not use Customer Content to train AI systems or models, except where explicitly agreed for specific functionality enhancements.</p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">9. Warranties and Disclaimers</h2>
          <h3 className="text-[15px] font-semibold text-white mb-2">9.1 Limited Warranty</h3>
          <p className="mb-6">We warrant that the ClawdFace Services will materially conform to the Documentation when used in accordance with these Terms.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">9.2 Disclaimer</h3>
          <p className="mb-8 font-semibold">EXCEPT AS EXPRESSLY PROVIDED, THE SERVICES ARE PROVIDED "AS IS" WITHOUT WARRANTIES OF ANY KIND. WE DISCLAIM ALL IMPLIED WARRANTIES INCLUDING MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT.</p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">10. Limitation of Liability</h2>
          <h3 className="text-[15px] font-semibold text-white mb-2">10.1 Damages Limitation</h3>
          <p className="mb-6 font-semibold">TO THE FULLEST EXTENT PERMITTED BY LAW, WE SHALL NOT BE LIABLE FOR INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES, INCLUDING LOST PROFITS OR DATA.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">10.2 Liability Cap</h3>
          <p className="mb-8">Our total liability shall not exceed the greater of: (i) amounts paid by you in the three months preceding the claim, (ii) $100, or (iii) applicable statutory remedies.</p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">11. Indemnification</h2>
          <h3 className="text-[15px] font-semibold text-white mb-2">11.1 Your Indemnification</h3>
          <p className="mb-6">You will indemnify us against claims arising from: (a) your use of the services contrary to these Terms, (b) your Customer Content, (c) violation of third-party rights, or (d) your breach of these Terms.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">11.2 Our Indemnification</h3>
          <p className="mb-8">We will indemnify you against third-party claims that your authorized use of the ClawdFace Solution infringes intellectual property rights, subject to the limitations and procedures outlined in these Terms.</p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">12. Term and Termination</h2>
          <h3 className="text-[15px] font-semibold text-white mb-2">12.1 Term</h3>
          <p className="mb-6">These Terms commence when you accept them and continue until terminated in accordance with these Terms.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">12.2 Termination Rights</h3>
          <p className="mb-6">Either party may terminate for material breach that remains uncured after 30 days' written notice. We may also terminate immediately for certain violations or if required by law.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">12.3 Effect of Termination</h3>
          <p className="mb-8">Upon termination, your right to use the services ends, and we may delete associated data. Provisions that by their nature should survive will remain in effect.</p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">13. Dispute Resolution</h2>
          <h3 className="text-[15px] font-semibold text-white mb-2">13.1 Governing Law</h3>
          <p className="mb-6">These Terms are governed by the laws of New Jersey.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">13.2 Arbitration</h3>
          <p className="mb-6">Subject to certain exceptions, disputes will be resolved through binding arbitration rather than court proceedings. You waive rights to jury trial and class action participation unless you opt out within 30 days.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">13.3 Informal Resolution</h3>
          <p className="mb-8">Before initiating arbitration, parties must engage in good faith informal dispute resolution.</p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">14. General Provisions</h2>
          <h3 className="text-[15px] font-semibold text-white mb-2">14.1 Updates</h3>
          <p className="mb-6">We may update these Terms at any time. Material changes will be communicated to registered users at least 30 days in advance.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">14.2 Assignment</h3>
          <p className="mb-6">You may not assign these Terms without our consent. We may freely assign our rights and obligations.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">14.3 Severability</h3>
          <p className="mb-6">If any provision is found unenforceable, it will be modified to reflect the original intent, and the remainder of these Terms will remain in effect.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">14.4 Force Majeure</h3>
          <p className="mb-6">Neither party is liable for delays or failures due to circumstances beyond reasonable control.</p>

          <h3 className="text-[15px] font-semibold text-white mt-8 mb-2">14.5 Entire Agreement</h3>
          <p className="mb-8">These Terms constitute the entire agreement between us regarding the subject matter and supersede all prior agreements.</p>

          <h2 className="text-2xl font-bold text-white mt-12 mb-6">15. Contact Information</h2>
          <p className="mb-4">For questions about these Terms or to provide notice or intellectual property infringement claims:</p>
          <p className="mb-8">Email: <a className="text-brand-muted hover:text-[#6bd69e] hover:underline transition-colors" href="mailto:support@clawdface.io">support@clawdface.io</a></p>
          <p className="mb-8 mt-8 border-t border-white/10 pt-8 italic text-zinc-400">By using the ClawdFace Platform, you acknowledge that you have read, understood, and agree to be bound by these Terms of Service.</p>
        </div>
      </main>
      <Footer />
    </div>
  );
}
