const {randomShareCode, TestEmail} = require("../../support/e2e");

describe('Choose not to be a certificate provider', () => {
    it('can enter reference number to not be a certificate provider', () => {
        const shareCode = randomShareCode()

        cy.visit(`/fixtures/certificate-provider?redirect=/certificate-provider-enter-reference-number-opt-out&withShareCode=${shareCode}&email=${TestEmail}`)

        cy.checkA11yApp();

        cy.get('#f-reference-number').type(shareCode);
        cy.contains('Continue').click();

        cy.url().should('contain', '/confirm-you-do-not-want-to-be-a-certificate-provider')
    });
})
