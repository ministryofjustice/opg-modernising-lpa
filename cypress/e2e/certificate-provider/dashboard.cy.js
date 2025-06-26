import {randomShareCode, TestEmail} from "../../support/e2e.js";

describe('Dashboard', () => {
    context('confirmed identity', () => {
        it('shows the certificate provider card', () => {
            cy.visit('/fixtures/certificate-provider?redirect=/task-list&progress=confirmYourIdentity');

            cy.contains('li', 'Confirm your details').should('contain', 'Completed');
            cy.contains('li', 'Confirm your identity').should('contain', 'Completed');
            cy.contains('li', 'Provide your certificate').should('contain', 'Not started');

            cy.visit('/dashboard')

            cy.contains('I’m a certificate provider').click()
            cy.contains('Property and affairs');
            cy.contains('Sam Smith');
            cy.contains('strong', 'In progress');
            cy.contains('a', 'Go to task list')
        })
    })

    context('provided certificate', () => {
        it('does not show the certificate provider card', () => {
            cy.visit('/fixtures/certificate-provider?redirect=/task-list&progress=provideYourCertificate');

            cy.contains('li', 'Confirm your details').should('contain', 'Completed');
            cy.contains('li', 'Confirm your identity').should('contain', 'Completed');
            cy.contains('li', 'Provide your certificate').should('contain', 'Completed');

            cy.visit('/dashboard')

            cy.contains('I’m a certificate provider').should('not.exist')
        })
    })

    context('provided certificate but identity mismatch', () => {
        it('shows the certificate provider card', () => {
            cy.visit('/fixtures/certificate-provider?redirect=/task-list&progress=provideYourCertificate&idStatus=mismatch');

            cy.contains('li', 'Confirm your details').should('contain', 'Completed');
            cy.contains('li', 'Confirm your identity').should('contain', 'Pending');
            cy.contains('li', 'Provide your certificate').should('contain', 'Completed');

            cy.visit('/dashboard')

            cy.contains('I’m a certificate provider').click()
            cy.contains('Property and affairs');
            cy.contains('Sam Smith');
            cy.contains('strong', 'In progress');
            cy.contains('a', 'Go to task list')
        })
    })

    context('with existing LPAs and a certificate provider share code', () => {
        it('can add access to provide a certificate', () => {
            const randomCode = randomShareCode();
            cy.visit(`/fixtures/certificate-provider?redirect=&withShareCode=${randomCode}&progress=signedByDonor&email=${TestEmail}`);
            cy.visit(`/fixtures?redirect=/task-list&progress=provideYourDetails`);

            cy.contains('a', 'Make or add an LPA').click();

            cy.contains('a', 'Continue').click();

            cy.contains('label', 'I have a code inviting me to be a certificate provider').click();
            cy.url().should('contain', '/add-an-lpa');
            cy.checkA11yApp();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/certificate-provider-enter-access-code');
            cy.get('#f-access-code').invoke('val', randomCode);
            cy.checkA11yApp();

            cy.contains('button', 'Save and continue').click();
            cy.contains('We have identified your certificate provider access code').click();
        })
    })
})
