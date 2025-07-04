import { randomAccessCode, TestEmail } from "../../support/e2e.js";

describe('Dashboard', () => {
    context('with existing LPAs and a voucher share code', () => {
        it('can add access to vouch for a donor identity', () => {
            const randomCode = randomAccessCode();
            cy.visit(`/fixtures/voucher?redirect=&withAccessCode=${randomCode}&progress=&email=${TestEmail}`);
            cy.visit(`/fixtures?redirect=/task-list&progress=provideYourDetails`);

            cy.contains('a', 'Make or add an LPA').click();

            cy.contains('a', 'Continue').click();

            cy.contains('label', 'I have a code inviting me to verify someone’s identity').click();
            cy.url().should('contain', '/add-an-lpa');
            cy.checkA11yApp();
            cy.contains('button', 'Continue').click();

            cy.url().should('contain', '/voucher-enter-access-code');
            cy.get('#f-donor-last-name').type('Smith');
            cy.get('#f-access-code').invoke('val', randomCode);
            cy.checkA11yApp();

            cy.contains('button', 'Save and continue').click();
            cy.contains('Vouch for someone’s identity').click();
        })
    })

})
