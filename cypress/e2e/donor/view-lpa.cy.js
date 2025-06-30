describe('View LPA', () => {
    describe('when signed by donor', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/view-lpa&progress=signTheLpa&donor=signature-expired');
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
            cy.contains('As soon as the LPA registration process is complete (including when the donor still has mental capacity to make a particular decision)');
            cy.contains('Jointly and severally - attorneys can make decisions both on their own or together');
            cy.contains('All together, as soon as one of your original attorneys can no longer act');
            cy.contains('My attorneys must not sell my home unless, in my doctor’s opinion, I can no longer live independently');
        });

        it('contains the donor signature', () => {
            cy.contains('Signed by Sam Smith on: 1 January 2024');
            cy.contains('Witnessed by Charlie Cooper on: 1 January 2024');
        });

        it('does not contain other signatures', () => {
            cy.contains('Attorney signature').should('not.exist');
        });
    });

    describe('when signed by everyone', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/view-lpa&attorneys=trust-corporation&progress=statutoryWaitingPeriod&donor=signature-expired');
        });

        it('shows all signatures', () => {
            cy.contains('Signed by Sam Smith on: 1 January 2024');
            cy.contains('Witnessed by Charlie Cooper on: 1 January 2024');
            cy.contains('Signed by Charlie Cooper on: 4 January 2024');
            cy.contains('Signed by Jessie Jones on: 11 January 2024');
            cy.contains('Signed by Robin Redcar on: 11 January 2024');
            cy.contains('Signed by A Sign on: 16 January 2024');
            cy.contains('Signed by Blake Buckley on: 11 January 2024');
            cy.contains('Signed by Taylor Thompson on: 11 January 2024');
        });
    });
});
