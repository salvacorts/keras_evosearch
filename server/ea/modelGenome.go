package ea

import (
	"math/rand"
	pb "server/protobuf/api"

	"github.com/MaxHalford/eaopt"
	uuid "github.com/satori/go.uuid"
)

var (
	ActivationFuncs = []pb.ActivationFunc{
		pb.ActivationFunc_Relu,
		pb.ActivationFunc_Sigmoid,
		pb.ActivationFunc_Softmax,
		pb.ActivationFunc_Tanh,
	}

	Optimizers = []pb.Optimizer{
		pb.Optimizer_Adam,
		pb.Optimizer_RMSprop,
		pb.Optimizer_SGD,
	}

	Models_chan_to_evaluate = make(map[string](chan *pb.ModelParameters))
	Models_chan_evaluated   = make(map[string](chan *pb.ModelResults))

	MutationRate = 0.3
	CrossRate    = 0.3
)

type ModelGenome struct {
	pb.ModelParameters
}

func (m *ModelGenome) Evaluate() (float64, error) {
	Models_chan_to_evaluate[m.ModelId] <- &m.ModelParameters
	result := <-Models_chan_evaluated[m.ModelId]

	return 1.0 - float64(result.Recall), nil
}

// Mutate by randomly modifying the LR, Optimizer, ActivationFunc and Droput
// Also randomly add or delete up to a 25% of the number of neurons on each layer
func (m *ModelGenome) Mutate(rng *rand.Rand) {
	rndFloatInRange := func(min float32, max float32) float32 {
		return ((max - min) * rng.Float32()) + min
	}
	rndIntInRange := func(min int32, max int32) int32 {
		return ((max - min) * rng.Int31()) + min
	}

	if rng.Float64() < MutationRate {
		delta := rndFloatInRange(-0.05, 0.05)
		m.LearningRate += delta

		if m.LearningRate >= 1 || m.LearningRate <= 0 {
			m.LearningRate -= delta
		}
	}

	if rng.Float64() < MutationRate {
		m.Optimizer = Optimizers[rand.Intn(len(Optimizers))]
	}

	if rng.Float64() < MutationRate {
		m.ActivationFunc = ActivationFuncs[rand.Intn(len(ActivationFuncs))]
	}

	if rng.Float64() < MutationRate {
		m.Dropout = !m.Dropout
	}

	// Add/Substract the 25% of the neurons in each layer
	for _, layer := range m.Layers {
		if rng.Float64() < MutationRate {
			n := int32(float32(layer.NumNeurons) * 0.25)
			layer.NumNeurons += rndIntInRange(-n, n)
		}
	}

	if rng.Float64() < MutationRate {
		eaopt.MutPermute(Layers(m.Layers), 3, rng)
	}
}

// Swaps LR, ActivationFunc, Dropout, Optimizer and layer with a given probability
func (m *ModelGenome) Crossover(Y eaopt.Genome, rng *rand.Rand) {
	other := Y.(*ModelGenome)

	if rng.Float64() < CrossRate {
		m.Optimizer, other.Optimizer = other.Optimizer, m.Optimizer
	}

	if rng.Float64() < CrossRate {
		m.ActivationFunc, other.ActivationFunc = other.ActivationFunc, m.ActivationFunc
	}

	if rng.Float64() < CrossRate {
		m.Dropout, other.Dropout = other.Dropout, m.Dropout
	}

	if rng.Float64() < CrossRate {
		var n int = 2
		if len(m.Layers) < n || len(other.Layers) < n {
			n = 1
		}

		eaopt.CrossGNX(Layers(m.Layers), Layers(other.Layers), uint(n), rng)
	}
}

func (m *ModelGenome) Clone() eaopt.Genome {
	mNew := ModelGenome{
		pb.ModelParameters{
			ModelId:        m.ModelId,
			LearningRate:   m.LearningRate,
			Optimizer:      m.Optimizer,
			ActivationFunc: m.ActivationFunc,
			Layers:         make([]*pb.Layer, len(m.Layers)),
			Dropout:        m.Dropout,
		},
	}

	copy(mNew.Layers, m.Layers)

	return &mNew
}

func MakeModel(rng *rand.Rand) eaopt.Genome {

	min_layers, max_layes := 1, 5
	min_neurons, max_neurons := 2, 256

	m := ModelGenome{
		pb.ModelParameters{
			ModelId:        uuid.NewV4().String(),
			LearningRate:   rand.Float32() / 10,
			Optimizer:      Optimizers[rand.Intn(len(Optimizers))],
			ActivationFunc: ActivationFuncs[rand.Intn(len(ActivationFuncs))],
			Layers:         make([]*pb.Layer, rand.Intn(max_layes)+min_layers),
			Dropout:        rand.Intn(2) == 1,
		},
	}

	for i := range m.Layers {
		m.Layers[i] = &pb.Layer{
			NumNeurons: rand.Int31n(int32(max_neurons)) + int32(min_neurons),
		}
	}

	Models_chan_to_evaluate[m.ModelId] = make(chan *pb.ModelParameters, 1)
	Models_chan_evaluated[m.ModelId] = make(chan *pb.ModelResults, 1)

	return &m
}
