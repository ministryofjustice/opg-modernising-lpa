describe('View LPA', () => {
    describe('when signed by donor', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/view-lpa&progress=signTheLpa');
        });

        it('shows the actors', () => {
            cy.contains('Donor: Sam Smith');
            cy.contains('Certificate provider: Charlie Cooper');
            cy.contains('Attorney: Jessie Jones');
            cy.contains('Attorney: Robin Redcar');
            cy.contains('Replacement attorney: Blake Buckley');
            cy.contains('Replacement attorney: Taylor Thompson');
            cy.contains('Person to notify: Jordan Jefferson');
            cy.contains('Person to notify: Danni Davies');
        });

        it('shows the decisions', () => {
            cy.contains('Whether or not I have mental capacity to make a particular decision myself');
            cy.contains('Jointly and severally - your attorneys can make decisions both on their own or together');
            cy.contains('All together, as soon as one of your attorneys can no longer act. They will be able to make decisions jointly and severally with any attorney who is continuing to act.');
            cy.contains('My attorneys must not sell my home unless, in my doctor’s opinion, I can no longer live independently');
        });

        it('contains the donor signature', () => {
            cy.contains('Signed by Sam Smith on: 2 January 2023');
            cy.contains('Witnessed by Charlie Cooper on: 2 January 2023');
        });

        it('does not contain other signatures', () => {
            cy.contains('Attorney signature').should('not.exist');
        });
    });

    describe.only('when signed by everyone', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/view-lpa&attorneys=trust-corporation&progress=submitted');
        });

        it('shows all signatures', () => {
            cy.contains('Signed by Sam Smith on: 2 January 2023');
            cy.contains('Witnessed by Charlie Cooper on: 2 January 2023');
            cy.contains('Signed by Charlie Cooper on: 5 January 2023');
            cy.contains('Signed by Jessie Jones on: 12 January 2023');
            cy.contains('Signed by Robin Redcar on: 12 January 2023');
            cy.contains('Signed by A Sign on: 17 January 2023');
            cy.contains('Signed by Blake Buckley on: 12 January 2023');
            cy.contains('Signed by Taylor Thompson on: 12 January 2023');
        });
    });
});
