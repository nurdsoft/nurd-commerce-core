package service

// TODO commenting this for now, need to go back and refine unit tests
/*
func TestGetRatesEstimate(t *testing.T) {
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	tests := []struct {
		name         string
		mockResponse func(mockClient *client.MockClient)
		shipment     entities.Shipment
		expectError  bool
		expectedResp []entities.ShippingRate
	}{
		{
			name: "Successful GetRatesEstimate",
			mockResponse: func(mockClient *client.MockClient) {
				mockClient.EXPECT().
					Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, reqURL string, headers map[string]string, in, out interface{}) error {
						resp := []entities.ShippingRate{
							{
								CarrierID:      "carrier-1",
								ShippingAmount: entities.ShippingAmount{Amount: 10.0, Currency: "USD"},
							},
						}
						respBytes, _ := json.Marshal(resp)
						_ = json.Unmarshal(respBytes, out)
						return nil
					})
			},
			from: entities.ShippingAddress{
				Country: "US",
				Zip:     "12345",
				City:    "New York",
				State:   "NY",
			},
			to: entities.ShippingAddress{
				Country: "US",
				Zip:     "67890",
				City:    "Los Angeles",
				State:   "CA",
			},
			dimensions: entities.Dimensions{
				Length: decimal.NewFromInt(10),
				Width:  decimal.NewFromInt(5),
				Height: decimal.NewFromInt(4),
				Weight: decimal.NewFromInt(2),
			},
			expectError: false,
			expectedResp: []entities.EstimateRatesResponse{
				{
					CarrierID:      "carrier-1",
					ShippingAmount: entities.ShippingAmount{Amount: 10.0, Currency: "USD"},
				},
			},
		},
		{
			name: "Failed GetRatesEstimate",
			mockResponse: func(mockClient *client.MockClient) {
				mockClient.EXPECT().
					Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(assert.AnError)
			},

			from: entities.ShippingAddress{
				Country: "US",
				Zip:     "12345",
				City:    "New York",
				State:   "NY",
			},
			to: entities.ShippingAddress{
				Country: "US",
				Zip:     "67890",
				City:    "Los Angeles",
				State:   "CA",
			},
			dimensions: entities.Dimensions{
				Length: decimal.NewFromInt(10),
				Width:  decimal.NewFromInt(5),
				Height: decimal.NewFromInt(4),
				Weight: decimal.NewFromInt(2),
			},
			expectError:  true,
			expectedResp: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := client.NewMockClient(ctrl)
			tt.mockResponse(mockClient)

			svc := &service{
				httpClient: mockClient,
				config: config.Config{
					CarrierIds: "carrier-1,carrier-2",
				},
				logger: logger,
			}

			resp, err := svc.GetShippingRates(context.TODO(), tt.from, tt.to, tt.dimensions)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

*/
