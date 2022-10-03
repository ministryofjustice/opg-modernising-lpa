describe('Identity with Easy ID', () => {
    it("submits the completed LPA", () => {
        cy.visit('/testing-start?redirect=/id/easy-id');

        cy.contains('Hi Test Person');
    })
});
