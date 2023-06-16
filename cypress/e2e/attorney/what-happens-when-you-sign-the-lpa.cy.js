describe('What happens when you sign the LPA', () => {
    it('as a property and affairs attorney', () => {
        cy.visit('/testing-start?redirect=/attorney-what-happens-when-you-sign-the-lpa&completeLpa=1&asAttorney=1&signedByDonor=1&asCertificateProvider=certified&loginAs=attorney');

        cy.contains('h1', "What happens when you sign the LPA")
        cy.contains('p', "you’re officially saying that you want to be an attorney on")
        cy.contains('li', "make decisions about their money or property")
        cy.contains('strong', "should only do these things if the donor asks you to")

        cy.contains('Continue to signing page').click();

        cy.url().should('contain', '/attorney-sign');
    });

    it('as a personal welfare attorney', () => {
        cy.visit('/testing-start?redirect=/attorney-what-happens-when-you-sign-the-lpa&completeLpa=1&asAttorney=1&withType=hw&loginAs=attorney');

        cy.contains('p', "you’re officially saying that you want to be an attorney on")
        cy.contains('li', "their personal and medical care")
        cy.contains('strong', "cannot act on their behalf")
    });

    it('as a property and affairs replacement attorney', () => {
        cy.visit('/testing-start?redirect=/attorney-what-happens-when-you-sign-the-lpa&completeLpa=1&asReplacementAttorney=1');

        cy.contains('p', "you’re saying that you want to be a replacement attorney")
        cy.contains('li', "make decisions about their money or property")
        cy.contains('strong', "should only do these things if the donor asks you to")
    });

    it('as a personal welfare replacement attorney', () => {
        cy.visit('/testing-start?redirect=/attorney-what-happens-when-you-sign-the-lpa&completeLpa=1&asReplacementAttorney=1&withType=hw');

        cy.contains('p', "you’re saying that you want to be a replacement attorney ")
        cy.contains('li', "their personal and medical care")
        cy.contains('strong', "cannot act on their behalf")
    });
});
