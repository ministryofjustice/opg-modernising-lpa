const {randomShareCode, TestEmail} = require("../../support/e2e");

describe('Choose not to be a certificate provider', () => {
    describe('can enter reference number to not be a certificate provider', () => {
        it('when LPA has been signed and witnessed', () => {
            const shareCode = randomShareCode()
            cy.visit(`/fixtures/certificate-provider?redirect=/certificate-provider-enter-reference-number-opt-out&withShareCode=${shareCode}&email=${TestEmail}&progress=signedByDonor`)

            cy.checkA11yApp();

            cy.get('#f-reference-number').type(shareCode);
            cy.contains('Continue').click();

            cy.url().should('contain', '/confirm-you-do-not-want-to-be-a-certificate-provider')
            cy.checkA11yApp();

            cy.contains('Property and affairs')

            cy.contains('button', 'Confirm').click();

            cy.url().should('contain', '/you-have-decided-not-to-be-a-certificate-provider')
            cy.checkA11yApp();

            cy.contains('You have confirmed that you do not want to be Sam Smith’s certificate provider')
            cy.contains('We have let Sam know about your decision.')
        });

        it('when LPA has not been signed and witnessed', () => {
            const shareCode = randomShareCode()
            cy.visit(`/fixtures/certificate-provider?redirect=/certificate-provider-enter-reference-number-opt-out&withShareCode=${shareCode}&email=${TestEmail}`)

            cy.checkA11yApp();

            cy.get('#f-reference-number').type(shareCode);
            cy.contains('Continue').click();

            cy.url().should('contain', '/confirm-you-do-not-want-to-be-a-certificate-provider')
            cy.checkA11yApp();

            cy.contains('Property and affairs')

            cy.contains('button', 'Confirm').click();

            cy.url().should('contain', '/you-have-decided-not-to-be-a-certificate-provider')
            cy.checkA11yApp();

            cy.contains('You have confirmed that you do not want to be Sam Smith’s certificate provider')
            cy.contains('We have let Sam know about your decision.')
        });
    })
})
