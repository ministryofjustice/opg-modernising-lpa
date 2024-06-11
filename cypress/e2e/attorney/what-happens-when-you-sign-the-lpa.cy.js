describe('What happens when you sign the LPA', () => {
    it('as a property and affairs attorney', () => {
        cy.visit('/fixtures/attorney?redirect=/what-happens-when-you-sign-the-lpa&progress=readTheLPA');

        cy.contains('h1', "What happens when you sign the LPA")
        cy.contains('p', "you’re officially saying that you want to be an attorney on")
        cy.contains('li', "make decisions about their money or property")
        cy.contains('strong', "should only do these things if the donor asks you to")

        cy.contains('Continue to signing page').click();

        cy.url().should('contain', '/sign');
    });

    it('as a personal welfare attorney', () => {
        cy.visit('/fixtures/attorney?redirect=/what-happens-when-you-sign-the-lpa&lpa-type=personal-welfare&progress=readTheLPA');

        cy.contains('p', "you’re officially saying that you want to be an attorney on")
        cy.contains('li', "their personal and medical care")
        cy.contains('strong', "cannot act on their behalf")
    });

    it('as a property and affairs replacement attorney', () => {
        cy.visit('/fixtures/attorney?redirect=/what-happens-when-you-sign-the-lpa&progress=readTheLPA&is-replacement=1');

        cy.contains('p', "you’re saying that you want to be a replacement attorney")
        cy.contains('li', "make decisions about their money or property")
        cy.contains('strong', "should only do these things if the donor asks you to")
    });

    it('as a personal welfare replacement attorney', () => {
        cy.visit('/fixtures/attorney?redirect=/what-happens-when-you-sign-the-lpa&lpa-type=personal-welfare&progress=readTheLPA&is-replacement=1');

        cy.contains('p', "you’re saying that you want to be a replacement attorney ")
        cy.contains('li', "their personal and medical care")
        cy.contains('strong', "cannot act on their behalf")
    });
});
