package query_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/golem-base/query"
	"github.com/stretchr/testify/require"
)

func pointerOf[T any](v T) *T {
	return &v
}

func TestParse(t *testing.T) {

	// fmt.Println(query.Parser.String())

	t.Run("quoted string", func(t *testing.T) {

		v, err := query.Parse(`name = "test\"2"`)
		require.NoError(t, err)

		require.Equal(
			t,
			&query.Expression{
				Or: &query.OrExpression{
					Left: &query.AndExpression{
						Left: &query.EqualExpr{
							Assign: &query.Equality{
								Var: "name",
								Value: &query.Value{
									String: pointerOf("test\"2"),
								},
							},
						},
					},
				},
			},
			v,
		)

	})

	t.Run("number", func(t *testing.T) {
		v, err := query.Parse(`name = 123`)
		require.NoError(t, err)

		require.Equal(
			t,
			&query.Expression{
				Or: &query.OrExpression{
					Left: &query.AndExpression{
						Left: &query.EqualExpr{
							Assign: &query.Equality{
								Var: "name",
								Value: &query.Value{
									Number: pointerOf(uint64(123)),
								},
							},
						},
					},
				},
			},
			v,
		)
	})

	t.Run("and", func(t *testing.T) {
		v, err := query.Parse(`name = 123 && name2 = "abc"`)
		require.NoError(t, err)

		require.Equal(t,
			&query.Expression{
				Or: &query.OrExpression{
					Left: &query.AndExpression{
						Left: &query.EqualExpr{
							Assign: &query.Equality{
								Var: "name",
								Value: &query.Value{
									Number: pointerOf(uint64(123)),
								},
							},
						},
						Right: []*query.AndRHS{
							{
								Op: "&&",
								Expr: &query.EqualExpr{
									Assign: &query.Equality{
										Var: "name2",
										Value: &query.Value{
											String: pointerOf("abc"),
										},
									},
								},
							},
						},
					},
				},
			},
			v,
		)
	})

	t.Run("or", func(t *testing.T) {
		v, err := query.Parse(`name = 123 || name2 = "abc"`)
		require.NoError(t, err)

		require.Equal(t,
			&query.Expression{
				Or: &query.OrExpression{
					Left: &query.AndExpression{
						Left: &query.EqualExpr{
							Assign: &query.Equality{
								Var: "name",
								Value: &query.Value{
									Number: pointerOf(uint64(123)),
								},
							},
						},
					},
					Right: []*query.OrRHS{
						{
							Op: "||",
							Expr: &query.AndExpression{
								Left: &query.EqualExpr{
									Assign: &query.Equality{
										Var: "name2",
										Value: &query.Value{
											String: pointerOf("abc"),
										},
									},
								},
							},
						},
					},
				},
			},
			v,
		)
	})

	t.Run("parentheses", func(t *testing.T) {
		v, err := query.Parse(`(name = 123 || name2 = "abc") && name3 = "def" || name4 = 456`)
		require.NoError(t, err)

		require.Equal(t,
			&query.Expression{
				Or: &query.OrExpression{
					Left: &query.AndExpression{
						Left: &query.EqualExpr{
							Paren: &query.Expression{
								Or: &query.OrExpression{
									Left: &query.AndExpression{
										Left: &query.EqualExpr{
											Assign: &query.Equality{
												Var: "name",
												Value: &query.Value{
													Number: pointerOf(uint64(123)),
												},
											},
										},
									},
									Right: []*query.OrRHS{
										{
											Op: "||",
											Expr: &query.AndExpression{
												Left: &query.EqualExpr{
													Assign: &query.Equality{
														Var: "name2",
														Value: &query.Value{
															String: pointerOf("abc"),
														},
													},
												},
											},
										},
									},
								},
							},
						},
						Right: []*query.AndRHS{
							{
								Op: "&&",
								Expr: &query.EqualExpr{
									Assign: &query.Equality{
										Var: "name3",
										Value: &query.Value{
											String: pointerOf("def"),
										},
									},
								},
							},
						},
					},
					Right: []*query.OrRHS{
						{
							Op: "||",
							Expr: &query.AndExpression{
								Left: &query.EqualExpr{
									Assign: &query.Equality{
										Var: "name4",
										Value: &query.Value{
											Number: pointerOf(uint64(456)),
										},
									},
								},
							},
						},
					},
				},
			},
			v,
		)
	})

	t.Run("invalid expression", func(t *testing.T) {
		_, err := query.Parse(`key = 8e`)
		require.Error(t, err, `1:8: unexpected token "e"`)
	})

}
